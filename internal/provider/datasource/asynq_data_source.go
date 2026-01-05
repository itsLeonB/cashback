package datasource

import (
	"crypto/tls"

	"github.com/hibiken/asynq"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/meq"
)

func ProvideAsynq() meq.DB {
	return meq.NewAsynqDB(logger.Global, redisClientOpts())
}

func redisClientOpts() asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     config.Global.Valkey.Addr,
		Password: config.Global.Valkey.Password,
		DB:       config.Global.Valkey.Db,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
}
