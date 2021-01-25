package main

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"

	"github.com/rubenv/sql-migrate"
	"github.com/yelsukov/otus-ha/backend/conf"
	"github.com/yelsukov/otus-ha/backend/server"
)

var VERSION = "0.0.1"

func connectDb(cfg *conf.Config) (*sql.DB, error) {
	db, err := sql.Open(
		"mysql",
		cfg.DbUser+":"+cfg.DbPassword+"@tcp("+cfg.DbHost+":"+cfg.DbPort+")/"+cfg.DbName+"?parseTime=true",
	)
	if err != nil {
		return nil, err
	}

	for i := 0; i < 10; i++ {
		if err = db.Ping(); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.DbMaxOpenConn)
	db.SetMaxIdleConns(cfg.DbMaxIdleConn)
	db.SetConnMaxLifetime(cfg.DbConnMaxLife)

	return db, nil
}

func main() {
	log.Info("==================================================================")
	log.Info("=                  Running REST API Service                      =")
	log.Info("==================================================================")
	log.Infof("Version: %v ", VERSION)

	ctx, cancel := context.WithCancel(context.Background())

	config, err := conf.PopulateConfig()
	if err != nil {
		log.WithError(err).Fatal("failed to populate configuration")
	}

	// Init logger
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	if config.DebugMode {
		log.SetLevel(log.DebugLevel)
	}

	// Connect to db and implement migrations
	log.Info("connecting to db...")
	db, err := connectDb(config)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to DB")
	}
	defer func() {
		if err = db.Close(); err != nil {
			log.WithError(err).Error("failed to close db connection")
		}
	}()
	log.Info("successfully connected to db")

	log.Info("implementing migrations")
	migrations := &migrate.FileMigrationSource{Dir: config.DbMigrationsPath}
	n, err := migrate.Exec(db, "mysql", migrations, migrate.Up)
	if err != nil {
		log.WithError(err).Fatal("failed to implement migrations")
	}
	log.Infof("applied %d migrations!", n)

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

	log.Info("running http server...")
	s := server.NewServer(ctx, config, db)
	s.Serve()
}
