package saga

import (
	"context"
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/yelsukov/otus-ha/dialogue/domain/entities"
	"github.com/yelsukov/otus-ha/dialogue/saga/queues"
	"github.com/yelsukov/otus-ha/dialogue/saga/storages"
)

type Orchestrator struct {
	store  storages.Storage
	queue  queues.Queue
	active bool
}

func NewOrchestrator(store storages.Storage, queue queues.Queue) *Orchestrator {
	return &Orchestrator{store, queue, false}
}

func (o *Orchestrator) IsActive() bool {
	return o.active
}

func (o *Orchestrator) Start(ctx context.Context) {
	if o.active {
		log.Warn("saga orchestrator already active")
		return
	}

	if !o.store.Connected() {
		log.Warn("cannot start saga orchestrator due to unconnected storage")
		return
	}
	go func() {
		o.active = true
		o.queue.Listen(ctx, o.onEvent)
		o.active = false
	}()
	log.Warn("saga orchestrator has been started")
}

func (o *Orchestrator) ExecuteSaga(ctx context.Context, saga *entities.Saga) error {
	if !o.active {
		return errors.New("orchestrator not active")
	}
	if err := o.store.Save(ctx, saga); err != nil {
		return err
	}

	return o.queue.Publish(ctx, saga)
}

func (o *Orchestrator) commitSaga(saga *entities.Saga) {
	if err := o.store.Del(context.Background(), saga.Id); err != nil {
		log.WithError(err).Error("failed to finalize saga")
	}
}

func (o *Orchestrator) rollbackSaga(saga *entities.Saga) {
	log.Infof("rolling back saga %s", saga.Id)
	if saga.Compensate == nil {
		log.Infof("saga %s have no compensation", saga.Id)
		return
	}
	if err := saga.Compensate(saga); err != nil {
		log.WithError(err).Errorf("Failed to compensate local trx for saga %s", saga.Id)
		return
	}
	if err := o.store.Del(context.Background(), saga.Id); err != nil {
		log.WithError(err).Error("failed to finalize saga")
	}
}

func (o *Orchestrator) onEvent(event queues.SagaInboundMessage) {
	saga, err := o.store.Get(context.Background(), event.SagaId)
	if err != nil {
		log.WithError(err).Error("failed to read saga from store")
	}

	switch event.Status {
	case queues.StatusSuccess:
		o.commitSaga(&saga)
	case queues.StatusAbort:
		o.rollbackSaga(&saga)
	default:
		log.Error("Unknown saga's step status")
	}
}
