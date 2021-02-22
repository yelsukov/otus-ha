package main

import (
	"context"
	"github.com/yelsukov/otus-ha/news/heater"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/yelsukov/otus-ha/news/bus"
	"github.com/yelsukov/otus-ha/news/cache"
	"github.com/yelsukov/otus-ha/news/domain/entities"
	"github.com/yelsukov/otus-ha/news/processor"
	"github.com/yelsukov/otus-ha/news/server"
	"github.com/yelsukov/otus-ha/news/storages"
	"github.com/yelsukov/otus-ha/news/vars"
)

func establishDbConn(ctx context.Context, dsn string) (*mongo.Client, error) {
	dbCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	conn, err := mongo.Connect(dbCtx, options.Client().ApplyURI(dsn))
	if err != nil {
		return nil, err
	}

	pingCtx, pingCtxCancel := context.WithTimeout(ctx, time.Second)
	defer pingCtxCancel()
	for i := 0; i < 10; i++ {
		if err = conn.Ping(pingCtx, readpref.Primary()); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func initLogger(DebugMode bool) {
	log.Info("==================================================================")
	log.Info("=                      Running News Service                      =")
	log.Info("==================================================================")
	log.Infof("Version: %v ", vars.VERSION)

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
	if vars.TOKEN == "" {
		log.Fatal("auth token for interaction with backend service is empty")
	}

	cfg, err := PopulateConfig()
	if err != nil {
		log.WithError(err).Fatal("failed to populate configuration")
	}

	initLogger(cfg.DebugMode)

	ctx, cancel := context.WithCancel(context.Background())

	log.Info("connecting to db...")
	conn, err := establishDbConn(ctx, cfg.MongoDSN)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to DB")
	}
	defer func() {
		log.Info("disconnecting from the DB...")
		if err = conn.Disconnect(ctx); err != nil {
			log.WithError(err).Error("failed to close db connection")
		} else {
			log.Info("DB connection has been closed")
		}
	}()
	log.Info("successfully connected to db")

	log.Info("connecting to cache...", cfg.CacheDSN)
	cacheClient := cache.NewCache(ctx)
	err = cacheClient.Connect(cfg.CacheDSN, cfg.CachePassword)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to Cache")
	}
	defer func() {
		log.Info("disconnecting from the cache...")
		if err = cacheClient.Disconnect(); err != nil {
			log.WithError(err).Error("failed to disconnect from cache")
		} else {
			log.Info("cache connection has been closed")
		}
	}()
	log.Info("successfully connected to cache")

	// Create storages
	log.Info("getting storages for db " + cfg.DbName)
	db := conn.Database(cfg.DbName)
	followerStorage := storages.NewFollowerStorage(ctx, db, cacheClient)
	eventStorage := storages.NewEventStorage(ctx, db, cacheClient)

	log.Info("running cache heater")
	cacheHeater := heater.NewCacheHeater(followerStorage, eventStorage, cacheClient)
	log.Info("heating followers cache")
	cacheHeater.HeatFollowers()
	log.Info("cache heater has been run")

	busChan := make(chan *entities.Event, cfg.BusPartitions)

	log.Info("running events bus listener")
	busListener := bus.NewBusListener(ctx, cfg.BusDSN, cfg.BusTopic, busChan, cfg.BusPartitions)
	busListener.Listen()
	log.Info("bus listener has been started")

	log.Info("running processors manager")
	manager := processor.NewProcessorsManager(ctx, busChan, cacheClient, cacheHeater, followerStorage, eventStorage, cfg.BusPartitions)
	go manager.StartProcessing()
	log.Info("processor manager started")

	// Create the interruption channel end lock until it gets interruption signal from OS
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM)
	// Run routine for gracefully shut down
	go func() {
		sig := <-c
		log.Infof("received the %+v call, shutting down", sig)
		cancel()
		signal.Stop(c)
	}()

	log.Info("creating http server and endpoint...")
	s := server.NewServer(ctx, eventStorage)
	log.Info("running http server...")
	s.Serve(cfg.ServerPort)
}
