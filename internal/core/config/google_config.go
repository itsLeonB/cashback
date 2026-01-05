package config

type Google struct {
	ServiceAccount string `split_words:"true" required:"true"`
}

func (Google) Prefix() string {
	return "GOOGLE"
}
