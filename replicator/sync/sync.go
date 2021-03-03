package sync

import (
	"context"
	"sync"

	"github.com/siddontang/go-log/log"
	"github.com/siddontang/go-mysql/canal"
	"github.com/siddontang/go-mysql/mysql"

	"github.com/yelsukov/otus-ha/replicator/conf"
)

type Sync struct {
	tarantool TarantoolProvider
	canal     *canal.Canal

	masterCh chan masterTask
	syncCh   chan syncTask
	wg       *sync.WaitGroup

	dataDir string
}

type syncTask struct {
	action string
	space  string
	data   [][]interface{}
}

type masterTask struct {
	pos   mysql.Position
	force bool
}

type TarantoolProvider interface {
	Connect(cfg *conf.Config) (err error)
	Insert(space string, values ...interface{}) error
	Delete(space, index string, key interface{}) error
	Disconnect()
}

func NewSync(cfg *conf.Config, tt TarantoolProvider) (*Sync, error) {
	r := new(Sync)
	r.dataDir = cfg.DataDir
	r.tarantool = tt

	r.masterCh = make(chan masterTask, 4096)
	r.syncCh = make(chan syncTask, 4096)
	r.wg = &sync.WaitGroup{}

	var err error
	srcRegexp := make([]string, len(cfg.ReplicateTables), len(cfg.ReplicateTables))
	for i, table := range cfg.ReplicateTables {
		srcRegexp[i] = cfg.MysqlDbName + "\\." + table
	}

	if r.canal, err = canal.NewCanal(&canal.Config{
		Addr:              cfg.MysqlDsn,
		User:              cfg.MysqlUser,
		Password:          cfg.MysqlPass,
		ServerID:          cfg.MysqlServerId,
		Charset:           "utf8",
		Flavor:            "mysql",
		IncludeTableRegex: srcRegexp,
		Dump: canal.DumpConfig{
			ExecutionPath: "/usr/bin/mysqldump",
			DiscardErr:    false,
		},
	}); err != nil {
		return r, err
	}
	r.canal.AddDumpTables(cfg.MysqlDbName, cfg.ReplicateTables...)

	// We must use binlog full row image
	if err = r.canal.CheckBinlogRowImage("FULL"); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Sync) Run(ctx context.Context, workersCnt uint32) error {
	master, err := loadMasterInfo(r.dataDir)
	if err != nil {
		return err
	}

	// Run consumer for master info channel
	r.wg.Add(1)
	go runPositionSaver(master, r.masterCh, r.wg)

	// Run consumers for synchronization channel
	var i uint32 = 1
	for ; i <= workersCnt; i++ {
		log.Infof("starting worker #%d", i)
		r.wg.Add(1)
		go runWorker(ctx, r.tarantool, r.syncCh, r.wg, i)
	}

	// Set handler and run canal for reading of mysql binlog
	r.canal.SetEventHandler(&eventHandler{ctx: ctx, masterCh: r.masterCh, syncCh: r.syncCh})
	return r.canal.RunFrom(master.position())
}

func runWorker(ctx context.Context, tt TarantoolProvider, sync chan syncTask, wg *sync.WaitGroup, n uint32) {
	defer func() {
		wg.Done()
		log.Infof("worker #%d has been stopped", n)
	}()

	var err error
	for task := range sync {
		switch task.action {
		case canal.InsertAction:
			for i, row := range task.data {
				if err = tt.Insert(task.space, row...); err != nil {
					log.Errorf("failed to insert row #%d: %s, %+v", i, err.Error(), row)
				}
			}

		case canal.DeleteAction:
			for i, row := range task.data {
				if err = tt.Delete(task.space, "primary", row[0]); err != nil {
					log.Errorf("failed to delete row #%d: %s", i, err.Error())
				}
			}

		case canal.UpdateAction:
			log.Error("replication for `update` action not implemented")

		default:
			log.Error("invalid rows action " + task.action)
		}
	}
}

func runPositionSaver(mi *masterInfo, ch chan masterTask, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		log.Infof("master info saver has been stopped")
	}()

	var err error
	for task := range ch {
		if err = mi.save(task.pos, task.force); err != nil {
			log.Error("failed to save master position " + task.pos.String() + ", sync will be stopped: " + err.Error())
			continue
		}
	}
}

func (r *Sync) Close() {
	r.canal.Close()
	r.tarantool.Disconnect()
	log.Info("closing master & sync channels")
	close(r.masterCh)
	close(r.syncCh)
	log.Info("waiting for consumers stop")
	r.wg.Wait()
}
