package dto

import (
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
	NetBalance              decimal.Decimal         `json:"netBalance"`
	TotalLentToFriend       decimal.Decimal         `json:"totalLentToFriend"`
	TotalBorrowedFromFriend decimal.Decimal         `json:"totalBorrowedFromFriend"`
	TransactionHistory      []FriendTransactionItem `json:"transactionHistory"`
	CurrencyCode            string                  `json:"currencyCode"`
}

type FriendTransactionItem struct {
	BaseDTO
	Type           string          `json:"type"`
	Amount         decimal.Decimal `json:"amount"`
	TransferMethod string          `json:"transferMethod"`
	Description    string          `json:"description"`
}

type FriendDetailsResponse struct {
	Friend                   FriendDetails `json:"friend"`
	Balance                  FriendBalance `json:"balance"`
	RedirectToRealFriendship uuid.UUID     `json:"redirectToRealFriendship,omitzero"`
}
