package config

type Flag struct {
	SubscriptionPurchaseEnabled bool `split_words:"true"`
}

func (Flag) Prefix() string {
	return "FLAG"
}
