package sync

import (
	"context"
	"errors"

	"github.com/siddontang/go-log/log"
	"github.com/siddontang/go-mysql/canal"
	"github.com/siddontang/go-mysql/mysql"
)

type eventHandler struct {
	ctx context.Context
	mi  *masterInfo
	tt  *Tarantool

	canal.DummyEventHandler
}

func (h *eventHandler) OnRow(e *canal.RowsEvent) error {
	var err error
	switch e.Action {
	case canal.InsertAction:
		for _, row := range e.Rows {
			if err = h.tt.Insert(e.Table.Name, row...); err != nil {
				return err
			}
		}

	case canal.DeleteAction:
		for _, row := range e.Rows {
			if err = h.tt.Delete(e.Table.Name, "primary", row[0]); err != nil {
				return err
			}
		}

	case canal.UpdateAction:
		log.Error("replication for `update` action not implemented")

	default:
		err = errors.New("invalid rows action " + e.Action)
	}

	if err == nil {
		err = h.ctx.Err()
	}

	return err
}

// TODO save position via channel or at goroutine to release the event emitter
func (h *eventHandler) OnPosSynced(pos mysql.Position, _ mysql.GTIDSet, _ bool) error {
	strPos := pos.String()
	log.Info("saving position " + strPos)
	if err := h.mi.save(pos); err != nil {
		log.Error("failed to save master position " + strPos + ", sync will be stopped: " + err.Error())
		return err
	}
	log.Info("position " + strPos + " has been saved")

	return h.ctx.Err()
}
