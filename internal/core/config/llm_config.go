package config

type LLM struct {
	ApiKey  string `split_words:"true"`
	BaseUrl string `split_words:"true"`
	Model   string
}

func (LLM) Prefix() string {
	return "LLM"
}
