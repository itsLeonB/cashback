package provider

import (
	"errors"

	"github.com/itsLeonB/cashback/internal/core/config"
	adminConfig "github.com/itsLeonB/cashback/internal/core/config/admin"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/provider/admin"
)

type Providers struct {
	*DataSources
	*Repositories
	*CoreServices
	*Services

	// Admin
	AdminRepos    *admin.Repositories
	AdminServices *admin.Services
}

func (p *Providers) Shutdown() error {
	var errs error
	if e := p.DataSources.Shutdown(); e != nil {
		errs = errors.Join(errs, e)
	}
	if e := p.CoreServices.Shutdown(); e != nil {
		errs = errors.Join(errs, e)
	}
	if e := p.Services.Shutdown(); e != nil {
		errs = errors.Join(errs, e)
	}
	return errs
}

func All() (*Providers, error) {
	dataSources, err := ProvideDataSource()
	if err != nil {
		return nil, err
	}

	repos := ProvideRepositories(dataSources.Gorm)
	adminRepos := admin.ProvideRepositories(dataSources.Gorm)

	coreSvcs, err := ProvideCoreServices()
	if err != nil {
		if e := dataSources.Shutdown(); e != nil {
			logger.Error(e)
		}
		return nil, err
	}

	return &Providers{
		DataSources:   dataSources,
		Repositories:  repos,
		CoreServices:  coreSvcs,
		Services:      ProvideServices(repos, coreSvcs, config.Global.Auth, config.Global.App, config.Global.Push),
		AdminRepos:    adminRepos,
		AdminServices: admin.ProvideServices(adminRepos, adminConfig.Global),
	}, nil
}
