package admin

import (
	"github.com/itsLeonB/cashback/internal/provider"
	adminProvider "github.com/itsLeonB/cashback/internal/provider/admin"
	"github.com/itsLeonB/go-authkit/authgin"
)

type Handlers struct {
	Auth         AuthHandler
	Plan         PlanHandler
	PlanVersion  PlanVersionHandler
	Subscription SubscriptionHandler
	Profile      ProfileHandler
	Payment      PaymentHandler
}

func ProvideHandlers(adminServices *adminProvider.Services, adminRepos *adminProvider.Repositories, domainServices *provider.Services) *Handlers {
	return &Handlers{
		AuthHandler{stateless: authgin.NewStatelessHandler(adminServices.Kit), userRepo: adminRepos.User},
		PlanHandler{domainServices.Plan},
		PlanVersionHandler{domainServices.PlanVersion},
		SubscriptionHandler{domainServices.Subscription},
		ProfileHandler{domainServices.Profile},
		PaymentHandler{domainServices.Payment},
	}
}
