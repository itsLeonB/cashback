package config

type Payment struct {
	Gateway      string `default:"ipaymu"`
	ServerKey    string `split_words:"true" required:"true"`
	Va           string `split_words:"true"`
	BaseUrl      string `split_words:"true" default:"https://sandbox.ipaymu.com"`
	ReturnUrl    string `split_words:"true"`
	NotifyUrl    string `split_words:"true"`
	CancelUrl    string `split_words:"true"`
	NotifySecret string `split_words:"true"` // shared secret for callback validation
}

func (Payment) Prefix() string {
	return "PAYMENT"
}
