package admin

import (
	adminConfig "github.com/itsLeonB/cashback/internal/core/config/admin"
	"github.com/itsLeonB/cashback/internal/domain/service/admin"
	"github.com/itsLeonB/sekure"
)

type Services struct {
	Auth admin.AuthService
}

func ProvideServices(repos *Repositories, cfg adminConfig.Config) *Services {

	return &Services{
		admin.NewAuthService(
			repos.User,
			sekure.NewHashService(cfg.HashCost),
			sekure.NewJwtService(cfg.Issuer, cfg.SecretKey, cfg.TokenDuration),
		),
	}
}
