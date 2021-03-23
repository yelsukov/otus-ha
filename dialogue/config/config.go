package config

import (
	"errors"
	"os"
	"strconv"
)

const (
	defDbName      = "otus"
	defServiceName = "dialogue"
	defServiceId   = "dialogue"
	defServiceHost = "http://dialogue"
	defServicePort = "8081"
	defConsulDsn   = "consul"
)

// Config struct holds application's parameters
type Config struct {
	MongoDSN string
	DbName   string

	ServiceName string
	ServiceId   string
	ServiceHost string
	ServicePort string

	ConsulDsn string

	ZabbixName string
	ZabbixHost string
	ZabbixPort int

	RedisDsn  string
	RedisPass string

	KafkaBrokers string

	DebugMode bool
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
		cfg.DbName = defDbName
	}

	if cfg.ServiceName, exist = os.LookupEnv("SERVICE_NAME"); !exist {
		cfg.ServiceName = defServiceName
	}
	if cfg.ServiceId, exist = os.LookupEnv("SERVICE_ID"); !exist {
		cfg.ServiceId = defServiceId
	}
	if cfg.ServiceHost, exist = os.LookupEnv("SERVICE_HOST"); !exist {
		cfg.ServiceHost = defServiceHost
	}
	if cfg.ServicePort, exist = os.LookupEnv("SERVICE_PORT"); !exist {
		cfg.ServicePort = defServicePort
	}

	if cfg.ConsulDsn, exist = os.LookupEnv("CONSUL_DSN"); !exist {
		cfg.ConsulDsn = defConsulDsn
	}

	cfg.ZabbixName = os.Getenv("ZBX_NAME")
	cfg.ZabbixHost = os.Getenv("ZBX_HOST")
	if zbxPort := os.Getenv("ZBX_PORT"); zbxPort != "" {
		cfg.ZabbixPort, _ = strconv.Atoi(zbxPort)
	}

	tmp, exist := os.LookupEnv("DEBUG")
	cfg.DebugMode = exist && tmp == "true"

	return &cfg, nil
}
