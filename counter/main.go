package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/yelsukov/otus-ha/counter/app"
	"github.com/yelsukov/otus-ha/counter/config"
	"github.com/yelsukov/otus-ha/counter/queues/kafka"
	"github.com/yelsukov/otus-ha/counter/storages/redis"
)

func initLogger(DebugMode bool) {
	log.Info("==================================================================")
	log.Info("=                    Running Counter Service                     =")
	log.Info("==================================================================")

	// Init logger
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	if DebugMode {
		log.SetLevel(log.DebugLevel)
	}

}

func main() {
	cfg, err := config.PopulateConfig()
	if err != nil {
		log.WithError(err).Error("failed to populate configuration")
		return
	}

	initLogger(cfg.DebugMode)

	ctx, cancel := context.WithCancel(context.Background())

	log.Info("connecting to db...")
	storage := redis.New()
	if err = storage.Connect(ctx, cfg.RedisDsn, cfg.RedisPass); err != nil {
		log.WithError(err).Error("failed to connect to Storage")
		return
	}
	log.Info("successfully connected to db")

	service := app.New(
		kafka.NewProducer(cfg.QueueDsn, cfg.DialogueTopic),
		kafka.NewConsumer(cfg.QueueDsn, cfg.CounterTopic),
		storage,
	)

	// Create the interruption channel end lock until it gets interruption signal from OS
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM)
	// Run routine for gracefully shut down
	go func(s *app.App) {
		sig := <-c
		log.Infof("received the %+v call, shutting down", sig)
		signal.Stop(c)

		s.Stop()
	}(service)

	// Lock until server shutdown
	log.Info("application ready and listening for events")
	if err = service.Run(ctx); err != nil {
		log.WithError(err).Error("fail on application runtime")
		service.Stop()
	}
	// cancel the base context
	cancel()
	log.Info("Shutdown completed!")
}
