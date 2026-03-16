package config

type Flag struct {
	SubscriptionPurchaseEnabled bool   `split_words:"true"`
	ClientKey                   string `split_words:"true"`
}

func (Flag) Prefix() string {
	return "FLAG"
}
