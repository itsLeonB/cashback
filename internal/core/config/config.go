package config

import (
	"errors"

	"github.com/itsLeonB/ungerr"
	"github.com/kelseyhightower/envconfig"
)

// type configurable interface {
// 	Prefix() string
// }

type Config struct {
	App
	Auth
	DB
	Google
	LLM
	Mail
	OAuthProviders
	Storage
	Valkey
}

var Global *Config

func Load() error {
	var errs error

	var app App
	if err := envconfig.Process("APP", &app); err != nil {
		errs = errors.Join(errs, err)
	}

	var valkey Valkey
	if err := envconfig.Process("VALKEY", &valkey); err != nil {
		errs = errors.Join(errs, err)
	}

	var storage Storage
	if err := envconfig.Process("STORAGE", &storage); err != nil {
		errs = errors.Join(errs, err)
	}

	var mail Mail
	if err := envconfig.Process("MAIL", &mail); err != nil {
		errs = errors.Join(errs, err)
	}

	oAuthProviders, err := loadOAuthProviderConfig()
	if err != nil {
		errs = errors.Join(errs, err)
	}

	var auth Auth
	if err = envconfig.Process("AUTH", &auth); err != nil {
		errs = errors.Join(errs, err)
	}

	var llm LLM
	if err = envconfig.Process("LLM", &llm); err != nil {
		errs = errors.Join(errs, err)
	}

	var db DB
	if err = envconfig.Process("DB", &db); err != nil {
		errs = errors.Join(errs, err)
	}

	var google Google
	if err = envconfig.Process("GOOGLE", &google); err != nil {
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		return ungerr.Wrap(errs, "error loading config")
	}

	Global = &Config{
		app,
		auth,
		db,
		google,
		llm,
		mail,
		oAuthProviders,
		storage,
		valkey,
	}

	return nil
}
