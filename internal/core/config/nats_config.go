package config

type NATS struct {
	Url string `required:"true"`
}

func (NATS) Prefix() string {
	return "NATS"
}
