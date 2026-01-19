package config

import (
	"context"

	"cloud.google.com/go/storage"
	vision "cloud.google.com/go/vision/apiv1"
	"github.com/itsLeonB/ungerr"
	"golang.org/x/oauth2/google"
)

type Google struct {
	ServiceAccount string `split_words:"true" required:"true"`
}

func (Google) Prefix() string {
	return "GOOGLE"
}

func (g *Google) LoadCredentials() (*google.Credentials, error) {
	scopes := append(vision.DefaultAuthScopes(), storage.ScopeFullControl)
	creds, err := google.CredentialsFromJSON(context.Background(), []byte(g.ServiceAccount), scopes...)
	if err != nil {
		return nil, ungerr.Wrap(err, "error parsing google credentials")
	}
	return creds, nil
}
