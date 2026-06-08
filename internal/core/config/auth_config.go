package config

import "time"

type Auth struct {
	SecretKey     string        `split_words:"true" default:"thisissecret"`
	TokenDuration time.Duration `split_words:"true" default:"24h"`
	Issuer        string        `default:"cashback"`
	HashCost      int           `split_words:"true" default:"10"`
}

func (Auth) Prefix() string {
	return "AUTH"
}
