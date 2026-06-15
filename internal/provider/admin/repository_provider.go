package admin

import (
	"github.com/itsLeonB/cashback/internal/domain/entity/admin"
	"github.com/itsLeonB/go-crud"
	"gorm.io/gorm"
)

type Repositories struct {
	Transactor crud.Transactor
	User       crud.Repository[admin.User]
}

func ProvideRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Transactor: crud.NewTransactor(db),
		User:       crud.NewRepository[admin.User](db),
	}
}
