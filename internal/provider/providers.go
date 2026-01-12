package provider

import (
	"errors"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
)

type Providers struct {
	*DataSources
	*Queues
	*Repositories
	*CoreServices
	*Services
}

func (p *Providers) Shutdown() error {
	var errs error
	if e := p.DataSources.Shutdown(); e != nil {
		errs = errors.Join(errs, e)
	}
	if e := p.Queues.Shutdown(); e != nil {
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

	queues, err := ProvideQueues(config.Global.Valkey)
	if err != nil {
		if e := dataSources.Shutdown(); e != nil {
			logger.Error(e)
		}
		return nil, err
	}
	repos := ProvideRepositories(dataSources)

	coreSvcs, err := ProvideCoreServices()
	if err != nil {
		if e := dataSources.Shutdown(); e != nil {
			logger.Error(e)
		}
		if e := queues.Shutdown(); e != nil {
			logger.Error(e)
		}
		return nil, err
	}

	return &Providers{
		DataSources:  dataSources,
		Queues:       queues,
		Repositories: repos,
		CoreServices: coreSvcs,
		Services:     ProvideServices(repos, coreSvcs, queues, config.Global.Auth, config.Global.App),
	}, nil
}
