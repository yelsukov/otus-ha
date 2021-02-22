package main

import (
	"errors"
	"os"
	"runtime"
	"strconv"
)

const (
	defaultServerPort = "8082"
	defaultDbName     = "otus_events"
	defaultBusTopic   = "user.events"
)

// Config struct holds application's parameters
type Config struct {
	BusTopic      string
	BusDSN        string
	BusPartitions int
	MongoDSN      string
	DbName        string
	ServerPort    string
	DebugMode     bool
	CacheDSN      string
	CachePassword string
}

func PopulateConfig() (*Config, error) {
	var (
		cfg   Config
		exist bool
		err   error
	)

	if cfg.BusDSN, exist = os.LookupEnv("BUS_DSN"); !exist {
		return nil, errors.New("ENV `BUS_DSN` should be specified")
	}
	if cfg.BusTopic, exist = os.LookupEnv("BUS_TOPIC"); !exist {
		cfg.BusTopic = defaultBusTopic
	}
	if qty, exist := os.LookupEnv("BUS_PARTITIONS"); exist {
		cfg.BusPartitions, err = strconv.Atoi(qty)
		if err != nil {
			return nil, err
		}
	} else {
		cfg.BusPartitions = runtime.NumCPU() / 3
	}

	if cfg.CacheDSN, exist = os.LookupEnv("CACHE_DSN"); !exist {
		return nil, errors.New("ENV `CACHE_DSN` should be specified")
	}
	cfg.CachePassword = os.Getenv("CACHE_PASSWORD")

	if cfg.MongoDSN, exist = os.LookupEnv("MONGO_DSN"); !exist {
		return nil, errors.New("ENV `MONGO_DSN` should be specified")
	}
	if cfg.DbName, exist = os.LookupEnv("DB_NAME"); !exist {
		cfg.DbName = defaultDbName
	}

	if cfg.ServerPort, exist = os.LookupEnv("SERVER_PORT"); !exist {
		cfg.ServerPort = defaultServerPort
	}

	tmp, exist := os.LookupEnv("DEBUG")
	cfg.DebugMode = exist && tmp == "true"

	return &cfg, err
}
