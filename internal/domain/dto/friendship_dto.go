package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/shopspring/decimal"
)

type NewAnonymousFriendshipRequest struct {
	ProfileID uuid.UUID `json:"-" binding:"-"`
	Name      string    `json:"name" binding:"required,min=3"`
}

type FriendshipResponse struct {
	BaseDTO
	Type          users.FriendshipType `json:"type"`
	ProfileID     uuid.UUID            `json:"profileId"`
	ProfileName   string               `json:"profileName"`
	ProfileAvatar string               `json:"profileAvatar"`
}

type FriendshipWithProfile struct {
	Friendship    FriendshipResponse
	UserProfile   ProfileResponse
	FriendProfile ProfileResponse
}

type FriendDetails struct {
	BaseDTO
	ProfileID  uuid.UUID            `json:"profileId"`
	Name       string               `json:"name"`
	Type       users.FriendshipType `json:"type"`
	Email      string               `json:"email,omitempty"`
	Phone      string               `json:"phone,omitempty"`
	Avatar     string               `json:"avatar,omitempty"`
	ProfileID1 uuid.UUID            `json:"profileId1"`
	ProfileID2 uuid.UUID            `json:"profileId2"`
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
	Friend                   FriendDetails             `json:"friend"`
	Balance                  FriendBalance             `json:"balance"`
	Transactions             []DebtTransactionResponse `json:"transactions"`
	Stats                    FriendStats               `json:"stats"`
	RedirectToRealFriendship uuid.UUID                 `json:"redirectToRealFriendship,omitzero"`
}
