package config

type Valkey struct {
	Addr      string `required:"true,min=3"`
	Password  string `required:"true"`
	Db        int
	EnableTls bool `split_words:"true"`
}

func (Valkey) Prefix() string {
	return "Valkey"
}
