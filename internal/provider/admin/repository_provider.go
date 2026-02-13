package admin

import (
	"github.com/itsLeonB/cashback/internal/domain/entity/admin"
	"github.com/itsLeonB/go-crud"
	"gorm.io/gorm"
)

type Repositories struct {
	User crud.Repository[admin.User]
}

func ProvideRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		crud.NewRepository[admin.User](db),
	}
}
