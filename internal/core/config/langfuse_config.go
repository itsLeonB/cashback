package config

type Langfuse struct {
	PublicKey string `required:"true" split_words:"true"`
	SecretKey string `required:"true" split_words:"true"`
	BaseUrl   string `split_words:"true" default:"https://cloud.langfuse.com"`
}

func (Langfuse) Prefix() string {
	return "LANGFUSE"
}
