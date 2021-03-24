package app

import (
	"context"
	log "github.com/sirupsen/logrus"

	"github.com/yelsukov/otus-ha/counter/queues"
	"github.com/yelsukov/otus-ha/counter/storages"
)

type App struct {
	producer queues.Producer
	consumer queues.Consumer
	storage  storages.Storage
	stopped  bool
}

func New(producer queues.Producer, consumer queues.Consumer, storage storages.Storage) App {
	return App{producer, consumer, storage, false}
}

func (a *App) Run(ctx context.Context) error {
	return a.consumer.Listen(ctx, a.consume)
}

func (a *App) Stop() {
	if a.stopped {
		return
	}
	a.stopped = true

	log.Info("closing producer")
	if err := a.producer.Close(); err != nil {
		log.WithError(err).Error("failed to close producer")
	}
	log.Info("closing consumer")
	if err := a.consumer.Close(); err != nil {
		log.WithError(err).Error("failed to close consumer")
	}
	log.Info("closing consumer")
	if err := a.storage.Close(); err != nil {
		log.WithError(err).Error("failed to close consumer")
	}
}
