package config

type Payment struct {
	Gateway   string `default:"midtrans"`
	ServerKey string `split_words:"true" required:"true"`
	Env       string `default:"sandbox"`
}

func (Payment) Prefix() string {
	return "PAYMENT"
}
