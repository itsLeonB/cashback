package admin

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
)

type User struct {
	crud.BaseEntity
	Email    string
	Password string
}

func (User) TableName() string {
	return "admin_users"
}

type Role struct {
	crud.BaseEntity
	Name string
}

func (Role) TableName() string {
	return "admin_roles"
}

type UserRole struct {
	crud.BaseEntity
	UserID uuid.UUID
	RoleID uuid.UUID
}

func (UserRole) TableName() string {
	return "admin_users_roles"
}
