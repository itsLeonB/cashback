package config

import (
	"context"
	"sync"

	"github.com/itsLeonB/ungerr"
	"golang.org/x/oauth2/google"
)

type Google struct {
	ServiceAccount string `split_words:"true" required:"true"`
}

func (Google) Prefix() string {
	return "GOOGLE"
}

var (
	googleCreds     *google.Credentials
	googleCredsOnce sync.Once
)

func LoadGoogleCredentials() (*google.Credentials, error) {
	var err error
	googleCredsOnce.Do(func() {
		creds, e := google.CredentialsFromJSON(context.Background(), []byte(Global.Google.ServiceAccount))
		if err != nil {
			err = ungerr.Wrap(e, "error parsing google credentials")
			return
		}

		googleCreds = creds
	})
	return googleCreds, err
}
