package sync

import (
	"github.com/siddontang/go-log/log"
	"github.com/tarantool/go-tarantool"

	"github.com/yelsukov/otus-ha/replicator/conf"
)

type Tarantool struct {
	conn *tarantool.Connection
}

func (t *Tarantool) Connect(cfg *conf.Config) (err error) {
	if t.conn, err = tarantool.Connect(cfg.TaranDsn, tarantool.Opts{
		Timeout:       cfg.TaranTimeout,
		Reconnect:     cfg.TaranReconnect, // interval
		MaxReconnects: cfg.TaranReconAttempts,
		User:          cfg.TaranUser,
		Pass:          cfg.TaranPassword,
	}); err == nil {
		_, err = t.conn.Ping()
	}

	return
}

func (t *Tarantool) Insert(space string, values ...interface{}) error {
	_, err := t.conn.Insert(space, values)
	return err
}

func (t *Tarantool) Delete(space, index string, key interface{}) error {
	_, err := t.conn.Delete(space, index, key)
	return err
}

func (t *Tarantool) Disconnect() {
	log.Info("disconnection from tarantool")
	if err := t.conn.Close(); err != nil {
		log.Error("failed to close tarantool connection: " + err.Error())
	} else {
		log.Info("successfully disconnected from tarantool")
	}
}
