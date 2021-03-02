package conf

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultDbName = "otus"
)

// Config struct holds application's parameters
type Config struct {
	TaranDsn           string
	TaranTimeout       time.Duration
	TaranReconnect     time.Duration
	TaranReconAttempts uint
	TaranUser          string
	TaranPassword      string

	DataDir         string
	ReplicateTables []string

	MysqlDsn      string
	MysqlDbName   string
	MysqlUser     string
	MysqlPass     string
	MysqlServerId uint32
}

func PopulateConfig() (*Config, error) {
	var (
		cfg   Config
		exist bool
	)

	if cfg.TaranDsn, exist = os.LookupEnv("TARAN_DSN"); !exist {
		return nil, errors.New("ENV `TARAN_DSN` should be specified")
	}
	cfg.TaranTimeout = 3 * time.Second
	cfg.TaranReconnect = 1 * time.Second
	cfg.TaranReconAttempts = 5
	cfg.TaranUser = os.Getenv("TARAN_USER")
	cfg.TaranPassword = os.Getenv("TARAN_PASS")

	cfg.DataDir = "/var/data/replicator"

	tables, exist := os.LookupEnv("REPL_TABLES")
	if !exist {
		return nil, errors.New("ENV `REPL_TABLES` should be specified")
	}
	cfg.ReplicateTables = strings.Split(tables, ",")

	if cfg.MysqlDsn, exist = os.LookupEnv("MYSQL_DSN"); !exist {
		return nil, errors.New("ENV `MYSQL_DSN` should be specified")
	}
	if cfg.MysqlDbName, exist = os.LookupEnv("MYSQL_DBNAME"); !exist {
		cfg.MysqlDbName = defaultDbName
	}
	if cfg.MysqlUser, exist = os.LookupEnv("MYSQL_USER"); !exist {
		return nil, errors.New("ENV `MYSQL_USER` should be specified")
	}
	if cfg.MysqlPass, exist = os.LookupEnv("MYSQL_PASS"); !exist {
		return nil, errors.New("ENV `MYSQL_PASS` should be specified")
	}
	if tmp, exist := os.LookupEnv("MYSQL_SERVER_ID"); !exist {
		cfg.MysqlServerId = 1
	} else {
		sid, err := strconv.Atoi(tmp)
		if err != nil {
			return nil, err
		}
		cfg.MysqlServerId = uint32(sid)
	}

	return &cfg, nil
}
