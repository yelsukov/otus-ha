package main

import (
	"context"
	"github.com/yelsukov/otus-ha/dialogue/server"
	"github.com/yelsukov/otus-ha/dialogue/server/endpoints"
	"github.com/yelsukov/otus-ha/dialogue/storages"
	"github.com/yelsukov/otus-ha/dialogue/vars"

	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	log "github.com/sirupsen/logrus"
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

func main() {
	log.Info("==================================================================")
	log.Info("=                    Running Dialogue Service                    =")
	log.Info("==================================================================")
	log.Infof("Version: %v ", vars.VERSION)

	// Init logger
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	if vars.TOKEN == "" {
		log.Fatal("auth token for interaction with backend service is empty")
	}

	cfg, err := PopulateConfig()
	if err != nil {
		log.WithError(err).Fatal("failed to populate configuration")
	}
	if cfg.DebugMode {
		log.SetLevel(log.DebugLevel)
	}

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
	db := conn.Database(cfg.DbName)
	log.Info("successfully connected to db")

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
	s := server.NewServer(ctx, cfg.ServerPort)
	s.MountRoutes("/chats", endpoints.GetChatsRoutes(storages.NewChatStorage(ctx, db)))
	s.MountRoutes("/messages", endpoints.GetMessagesRoutes(storages.NewMessageStorage(ctx, db)))
	log.Info("running http server...")
	s.Serve()

}
