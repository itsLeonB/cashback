package datasource

import (
	"crypto/tls"

	"github.com/hibiken/asynq"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/meq"
)

func ProvideAsynq(cfg config.Valkey) meq.DB {
	return meq.NewAsynqDB(logger.Global, redisClientOpts(cfg))
}

func redisClientOpts(cfg config.Valkey) asynq.RedisClientOpt {
	opt := asynq.RedisClientOpt{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.Db,
	}

	if cfg.EnableTls {
		opt.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	return opt
}
