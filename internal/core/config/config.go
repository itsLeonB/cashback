package config

import (
	"errors"

	"github.com/itsLeonB/cashback/internal/core/config/admin"
	"github.com/itsLeonB/ungerr"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2/google"
)

// type configurable interface {
// 	Prefix() string
// }

type Config struct {
	App
	Auth
	DB
	LLM
	Mail
	OAuthProviders
	Push
	Valkey
	GoogleCreds *google.Credentials
	Payment
	Flag
	OTel
}

var Global *Config

func Load() error {
	var errs error

	if err := admin.Load(); err != nil {
		errs = errors.Join(errs, err)
	}

	var app App
	if err := envconfig.Process("APP", &app); err != nil {
		errs = errors.Join(errs, err)
	}

	var valkey Valkey
	if err := envconfig.Process("VALKEY", &valkey); err != nil {
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

	var push Push
	if err = envconfig.Process(push.Prefix(), &push); err != nil {
		errs = errors.Join(errs, err)
	}

	var google Google
	if err = envconfig.Process(google.Prefix(), &google); err != nil {
		errs = errors.Join(errs, err)
	}
	googleCreds, err := google.LoadCredentials()
	if err != nil {
		errs = errors.Join(errs, err)
	}

	var payment Payment
	if err = envconfig.Process(payment.Prefix(), &payment); err != nil {
		errs = errors.Join(errs, err)
	}

	var flag Flag
	if err = envconfig.Process(flag.Prefix(), &flag); err != nil {
		errs = errors.Join(errs, err)
	}

	var otel OTel
	if err = envconfig.Process(otel.Prefix(), &otel); err != nil {
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		return ungerr.Wrap(errs, "error loading config")
	}

	Global = &Config{
		app,
		auth,
		db,
		llm,
		mail,
		oAuthProviders,
		push,
		valkey,
		googleCreds,
		payment,
		flag,
		otel,
	}

	return nil
}
