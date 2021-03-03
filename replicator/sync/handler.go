package sync

import (
	"context"
	"github.com/siddontang/go-mysql/canal"
	"github.com/siddontang/go-mysql/mysql"
)

type eventHandler struct {
	ctx context.Context

	masterCh chan masterTask
	syncCh   chan syncTask

	canal.DummyEventHandler
}

func (h *eventHandler) OnRow(e *canal.RowsEvent) error {
	h.syncCh <- syncTask{e.Action, e.Table.Name, e.Rows}
	return h.ctx.Err()
}

func (h *eventHandler) OnPosSynced(pos mysql.Position, _ mysql.GTIDSet, force bool) error {
	h.masterCh <- masterTask{pos, force}
	return h.ctx.Err()
}
