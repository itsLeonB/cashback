package admin

import "github.com/itsLeonB/cashback/internal/provider/admin"

type Handlers struct {
	Auth AuthHandler
}

func ProvideHandlers(services *admin.Services) *Handlers {
	return &Handlers{
		AuthHandler{services.Auth},
	}
}
