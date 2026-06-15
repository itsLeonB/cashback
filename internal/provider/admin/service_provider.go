package admin

import (
	"context"

	adminauth "github.com/itsLeonB/cashback/internal/adapters/auth/admin"
	adminConfig "github.com/itsLeonB/cashback/internal/core/config/admin"
	admin "github.com/itsLeonB/cashback/internal/domain/entity/admin"
	"github.com/itsLeonB/go-authkit"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type Services struct {
	Kit *authkit.AuthKit
}

func ProvideServices(repos *Repositories, cfg *adminConfig.Config) *Services {
	kit, err := authkit.New(authkit.Config{
		Stateless:   true,
		JWTIssuer:   cfg.Issuer,
		JWTSecret:   cfg.SecretKey,
		JWTDuration: cfg.TokenDuration,
	}, authkit.Deps{
		Tx:    repos.Transactor,
		Users: adminauth.NewUserStore(repos.User),
	}, authkit.Hooks{
		BeforeRegister: func(ctx context.Context, _ string) error {
			user, err := repos.User.FindFirst(ctx, crud.Specification[admin.User]{})
			if err != nil {
				return err
			}
			if !user.IsZero() {
				return ungerr.ForbiddenError("cannot register as there exists admin users")
			}
			return nil
		},
	})
	if err != nil {
		panic(err)
	}

	return &Services{Kit: kit}
}
