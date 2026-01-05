package users

import (
	"database/sql"

	"github.com/itsLeonB/go-crud"
)

type User struct {
	crud.BaseEntity
	Email               string
	Password            string
	Profile             UserProfile
	VerifiedAt          sql.NullTime
	PasswordResetTokens []PasswordResetToken `gorm:"foreignKey:UserID"`
}

func (u User) IsVerified() bool {
	return u.VerifiedAt.Valid && !u.VerifiedAt.Time.IsZero()
}
