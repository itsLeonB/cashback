package admin

import "time"

type Auth struct {
	SecretKey     string        `split_words:"true" default:"thisissecret"`
	TokenDuration time.Duration `split_words:"true" default:"1h"`
	Issuer        string        `default:"cashdash"`
	HashCost      int           `split_words:"true" default:"10"`
}

func (Auth) Prefix() string {
	return "ADMIN_AUTH"
}
