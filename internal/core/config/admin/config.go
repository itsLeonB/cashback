package admin

import (
	"github.com/itsLeonB/ungerr"
	"github.com/kelseyhightower/envconfig"
)

type loadable interface {
	Prefix() string
}

type Config struct {
	Auth
}

var Global *Config

func Load() error {
	auth, err := load[Auth]()
	if err != nil {
		return ungerr.Wrap(err, "error loading configs")
	}

	Global = &Config{
		auth,
	}

	return nil
}

func load[T loadable]() (T, error) {
	var cfg T
	err := envconfig.Process(cfg.Prefix(), &cfg)
	return cfg, err
}
