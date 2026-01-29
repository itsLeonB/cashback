package config

type Push struct {
	VapidPublicKey  string `split_words:"true" required:"true"`
	VapidPrivateKey string `split_words:"true" required:"true"`
	VapidSubject    string `split_words:"true" required:"true"`
	Enabled         bool   `default:"false"`
}

func (Push) Prefix() string {
	return "PUSH"
}
