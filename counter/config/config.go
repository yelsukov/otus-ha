package config

import (
	"errors"
	"os"
)

const (
	defaultCounterTopic  = "counterBus"
	defaultDialogueTopic = "dialogueBus"
)

// Config struct holds application's parameters
type Config struct {
	QueueDsn      string
	DialogueTopic string
	CounterTopic  string
	DebugMode     bool
	RedisDsn      string
	RedisPass     string
}

func PopulateConfig() (*Config, error) {
	var (
		cfg   Config
		exist bool
		err   error
	)

	if cfg.QueueDsn, exist = os.LookupEnv("QUEUE_DSN"); !exist {
		return nil, errors.New("ENV `QUEUE_DSN` should be specified")
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
	cfg.RedisPass = os.Getenv("REDIS_PASS")

	tmp, exist := os.LookupEnv("DEBUG")
	cfg.DebugMode = exist && tmp == "true"

	return &cfg, err
}
