package config

type OTel struct {
	Enabled              bool   `required:"true" default:"false"`
	ExporterOtlpEndpoint string `split_words:"true" required:"true"`
	ExporterOtlpInsecure bool   `split_words:"true" required:"true" default:"false"`
	ExporterOtlpHeaders  string `split_words:"true" required:"true"`
	ServiceName          string `split_words:"true" required:"true" default:"cashback"`
	ServiceInstanceId    string `split_words:"true" required:"true"`
}

func (o OTel) Prefix() string {
	return "OTEL"
}
