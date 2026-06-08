package config

type NATS struct {
	URL string `required:"true"`
}

func (NATS) Prefix() string {
	return "NATS"
}
