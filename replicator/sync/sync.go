package sync

import (
	"context"

	"github.com/siddontang/go-mysql/canal"

	"github.com/yelsukov/otus-ha/replicator/conf"
)

type Sync struct {
	tarantool *Tarantool
	canal     *canal.Canal
	dataDir   string
}

func NewSync(cfg *conf.Config) (*Sync, error) {
	r := new(Sync)
	r.dataDir = cfg.DataDir

	var err error
	r.tarantool = &Tarantool{}
	if err = r.tarantool.Connect(cfg); err != nil {
		return nil, err
	}

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

func (r *Sync) Run(ctx context.Context) error {
	master, err := loadMasterInfo(r.dataDir)
	if err != nil {
		return err
	}

	r.canal.SetEventHandler(&eventHandler{ctx: ctx, mi: master, tt: r.tarantool})

	pos := master.position()
	if err = r.canal.RunFrom(pos); err != nil {
		return err
	}

	return nil
}

func (r *Sync) Close() {
	r.canal.Close()
	r.tarantool.Disconnect()
}
