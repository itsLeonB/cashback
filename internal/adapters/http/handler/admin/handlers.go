package admin

import (
	"github.com/itsLeonB/cashback/internal/provider"
	"github.com/itsLeonB/cashback/internal/provider/admin"
)

type Handlers struct {
	Auth        AuthHandler
	Plan        PlanHandler
	PlanVersion PlanVersionHandler
}

func ProvideHandlers(services *admin.Services, domainServices *provider.Services) *Handlers {
	return &Handlers{
		AuthHandler{services.Auth},
		PlanHandler{domainServices.Plan},
		PlanVersionHandler{domainServices.PlanVersion},
	}
}
