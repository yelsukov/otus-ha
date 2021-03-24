package config

import (
	"errors"
	"os"
)

const (
	defaultCounterTopic  = "counter"
	defaultDialogueTopic = "dialogue"
)

// Config struct holds application's parameters
type Config struct {
	QueueDsn      string
	DialogueTopic string
	CounterTopic  string
	DebugMode     bool
	RedisDsn      string
	RedisPassword string
}

func PopulateConfig() (*Config, error) {
	var (
		cfg   Config
		exist bool
		err   error
	)

	if cfg.QueueDsn, exist = os.LookupEnv("BUS_DSN"); !exist {
		return nil, errors.New("ENV `BUS_DSN` should be specified")
	}

	if cfg.CounterTopic, exist = os.LookupEnv("CNT_TOPIC"); !exist {
		cfg.CounterTopic = defaultCounterTopic
	}
	if cfg.DialogueTopic, exist = os.LookupEnv("DLG_TOPIC"); !exist {
		cfg.CounterTopic = defaultDialogueTopic
	}

	if cfg.RedisDsn, exist = os.LookupEnv("REDIS_DSN"); !exist {
		return nil, errors.New("ENV `REDIS_DSN` should be specified")
	}
	cfg.RedisPassword = os.Getenv("REDIS_PASS")

	tmp, exist := os.LookupEnv("DEBUG")
	cfg.DebugMode = exist && tmp == "true"

	return &cfg, err
}
