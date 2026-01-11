package config

import "time"

const AppName = "Cashback"

type App struct {
	Env                       string        `default:"debug"`
	Port                      string        `default:"8080"`
	Timeout                   time.Duration `default:"10s"`
	ClientUrls                []string      `split_words:"true"`
	RegisterVerificationUrl   string        `split_words:"true"`
	ResetPasswordUrl          string        `split_words:"true"`
	BucketNameExpenseBill     string        `split_words:"true" required:"true"`
	BucketNameTransferMethods string        `split_words:"true" default:"transfer-methods"`
}

func (App) Prefix() string {
	return "APP"
}
