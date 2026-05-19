package config

type Payment struct {
	ServerKey     string `split_words:"true" required:"true"`
	WebhookSecret string `split_words:"true" required:"true"`
	SuccessURL    string `split_words:"true" required:"true"`
	CancelURL     string `split_words:"true" required:"true"`
}

func (Payment) Prefix() string {
	return "PAYMENT"
}
