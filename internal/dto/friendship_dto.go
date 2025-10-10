package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/shopspring/decimal"
)

type NewAnonymousFriendshipRequest struct {
	ProfileID uuid.UUID `json:"-" binding:"-"`
	Name      string    `json:"name" binding:"required,min=3"`
}

type FriendshipResponse struct {
	ID            uuid.UUID                  `json:"id"`
	Type          appconstant.FriendshipType `json:"type"`
	ProfileID     uuid.UUID                  `json:"profileId"`
	ProfileName   string                     `json:"profileName"`
	ProfileAvatar string                     `json:"profileAvatar"`
	CreatedAt     time.Time                  `json:"createdAt"`
	UpdatedAt     time.Time                  `json:"updatedAt"`
	DeletedAt     time.Time                  `json:"deletedAt,omitzero"`
}

type FriendshipWithProfile struct {
	Friendship    FriendshipResponse
	UserProfile   ProfileResponse
	FriendProfile ProfileResponse
}

type FriendDetails struct {
	ID        uuid.UUID                  `json:"id"`
	ProfileID uuid.UUID                  `json:"profileId"`
	Name      string                     `json:"name"`
	Type      appconstant.FriendshipType `json:"type"`
	Email     string                     `json:"email,omitempty"`
	Phone     string                     `json:"phone,omitempty"`
	Avatar    string                     `json:"avatar,omitempty"`
	CreatedAt time.Time                  `json:"createdAt"`
	UpdatedAt time.Time                  `json:"updatedAt"`
	DeletedAt time.Time                  `json:"deletedAt,omitzero"`
}

type FriendBalance struct {
	TotalOwedToYou decimal.Decimal `json:"totalOwedToYou"`
	TotalYouOwe    decimal.Decimal `json:"totalYouOwe"`
	NetBalance     decimal.Decimal `json:"netBalance"`
	CurrencyCode   string          `json:"currencyCode"`
}

type FriendStats struct {
	TotalTransactions        int             `json:"totalTransactions"`
	FirstTransactionDate     time.Time       `json:"firstTransactionDate"`
	LastTransactionDate      time.Time       `json:"lastTransactionDate"`
	MostUsedTransferMethod   string          `json:"mostUsedTransferMethod"`
	AverageTransactionAmount decimal.Decimal `json:"averageTransactionAmount"`
}

type FriendDetailsResponse struct {
	Friend       FriendDetails             `json:"friend"`
	Balance      FriendBalance             `json:"balance"`
	Transactions []DebtTransactionResponse `json:"transactions"`
	Stats        FriendStats               `json:"stats"`
}
