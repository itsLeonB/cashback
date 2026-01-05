package config

type Valkey struct {
	Addr     string `required:"true,min=3"`
	Password string `required:"true"`
	Db       int
}

func (Valkey) Prefix() string {
	return "Valkey"
}
