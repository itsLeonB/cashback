package config

import (
	"os"

	"github.com/itsLeonB/ungerr"
	"github.com/kelseyhightower/envconfig"
)

type Google struct {
	ServiceAccount string `split_words:"true" required:"true"`
}

func (Google) Prefix() string {
	return "GOOGLE"
}

func loadGoogleConfig() error {
	var google Google
	if err := envconfig.Process(google.Prefix(), &google); err != nil {
		return ungerr.Wrap(err, "error processing google config")
	}

	credsPath := "/tmp/gcp.json"
	if err := os.WriteFile(credsPath, []byte(google.ServiceAccount), 0600); err != nil {
		return ungerr.Wrap(err, "error writing service account JSON file")
	}

	if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credsPath); err != nil {
		return ungerr.Wrapf(err, "error setting GOOGLE_APP_CREDENTIALS to %s", credsPath)
	}

	return nil
}
