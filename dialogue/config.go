package main

import (
	"errors"
	"os"
)

const (
	defaultServerPort = "8081"
	defaultDbName     = "otus"
)

// Config struct holds application's parameters
type Config struct {
	MongoDSN   string
	DbName     string
	ServerPort string
	DebugMode  bool
}

func PopulateConfig() (*Config, error) {
	var (
		cfg   Config
		exist bool
	)

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

	return &cfg, nil
}
