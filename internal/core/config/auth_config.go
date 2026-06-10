package config

import "time"

type Auth struct {
	SecretKey             string        `split_words:"true" default:"thisissecret"`
	TokenDuration         time.Duration `split_words:"true" default:"24h"`
	RefreshTokenDuration  time.Duration `split_words:"true" default:"720h"`
	Issuer                string        `default:"cashback"`
	HashCost              int           `split_words:"true" default:"10"`
	StateStore            string        `split_words:"true" default:"inmemory"`
	CookieDomain          string        `split_words:"true" default:"localhost"`
	CookieSecure          bool          `split_words:"true" default:"false"`
}

func (Auth) Prefix() string {
	return "AUTH"
}
