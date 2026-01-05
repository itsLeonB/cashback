package users

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
)

type FriendshipType string

const (
	Real      FriendshipType = "REAL"
	Anonymous FriendshipType = "ANON"
)

type Friendship struct {
	crud.BaseEntity
	ProfileID1 uuid.UUID
	ProfileID2 uuid.UUID
	Type       FriendshipType
	Profile1   UserProfile `gorm:"foreignKey:ProfileID1"`
	Profile2   UserProfile `gorm:"foreignKey:ProfileID2"`
}

type FriendshipSpecification struct {
	crud.Specification[Friendship]
	Name string
}
