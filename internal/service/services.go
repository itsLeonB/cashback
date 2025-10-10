package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/orcashtrator/internal/domain/groupexpense"
	"github.com/itsLeonB/orcashtrator/internal/dto"
)

type AuthService interface {
	Register(ctx context.Context, request dto.RegisterRequest) (dto.RegisterResponse, error)
	InternalLogin(ctx context.Context, request dto.InternalLoginRequest) (dto.LoginResponse, error)
	VerifyToken(ctx context.Context, token string) (bool, map[string]any, error)
	GetOAuth2URL(ctx context.Context, provider string) (string, error)
	OAuth2Login(ctx context.Context, provider, code, state string) (dto.LoginResponse, error)
	VerifyRegistration(ctx context.Context, token string) (dto.LoginResponse, error)
	SendPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) (dto.LoginResponse, error)
}

type ProfileService interface {
	GetByID(ctx context.Context, id uuid.UUID) (dto.ProfileResponse, error)
	GetNames(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]string, error)
	Update(ctx context.Context, id uuid.UUID, name string) (dto.ProfileResponse, error)
	Search(ctx context.Context, profileID uuid.UUID, input string) ([]dto.ProfileResponse, error)
}

type FriendshipService interface {
	CreateAnonymous(ctx context.Context, request dto.NewAnonymousFriendshipRequest) (dto.FriendshipResponse, error)
	GetAll(ctx context.Context, profileID uuid.UUID) ([]dto.FriendshipResponse, error)
	IsFriends(ctx context.Context, profileID1, profileID2 uuid.UUID) (bool, bool, error)
}

type FriendshipRequestService interface {
	Send(ctx context.Context, userProfileID, friendProfileID uuid.UUID) error
	GetAllSent(ctx context.Context, userProfileID uuid.UUID) ([]dto.FriendshipRequestResponse, error)
	Cancel(ctx context.Context, userProfileID, reqID uuid.UUID) error
	GetAllReceived(ctx context.Context, userProfileID uuid.UUID) ([]dto.FriendshipRequestResponse, error)
	Ignore(ctx context.Context, userProfileID, reqID uuid.UUID) error
	Block(ctx context.Context, userProfileID, reqID uuid.UUID) error
	Unblock(ctx context.Context, userProfileID, reqID uuid.UUID) error
	Accept(ctx context.Context, userProfileID, reqID uuid.UUID) (dto.FriendshipResponse, error)
}

type FriendDetailsService interface {
	GetDetails(ctx context.Context, profileID, friendshipID uuid.UUID) (dto.FriendDetailsResponse, error)
}

type DebtService interface {
	RecordNewTransaction(ctx context.Context, request dto.NewDebtTransactionRequest) (dto.DebtTransactionResponse, error)
	GetTransactions(ctx context.Context, userProfileID uuid.UUID) ([]dto.DebtTransactionResponse, error)
	ProcessConfirmedGroupExpense(ctx context.Context, groupExpense groupexpense.GroupExpense) error
	GetAllByProfileIDs(ctx context.Context, userProfileID, friendProfileID uuid.UUID) ([]dto.DebtTransactionResponse, error)
}

type TransferMethodService interface {
	GetAll(ctx context.Context) ([]dto.TransferMethodResponse, error)
}

type GroupExpenseService interface {
	CreateDraft(ctx context.Context, request dto.NewGroupExpenseRequest) (dto.GroupExpenseResponse, error)
	GetAllCreated(ctx context.Context, userProfileID uuid.UUID) ([]dto.GroupExpenseResponse, error)
	GetDetails(ctx context.Context, id, userProfileID uuid.UUID) (dto.GroupExpenseResponse, error)
	ConfirmDraft(ctx context.Context, id, userProfileID uuid.UUID) (dto.GroupExpenseResponse, error)
}

type ExpenseItemService interface {
	Add(ctx context.Context, request dto.NewExpenseItemRequest) (dto.ExpenseItemResponse, error)
	GetDetails(ctx context.Context, groupExpenseID, expenseItemID, userProfileID uuid.UUID) (dto.ExpenseItemResponse, error)
	Update(ctx context.Context, request dto.UpdateExpenseItemRequest) (dto.ExpenseItemResponse, error)
	Remove(ctx context.Context, groupExpenseID, expenseItemID, userProfileID uuid.UUID) error
}

type OtherFeeService interface {
	Add(ctx context.Context, request dto.NewOtherFeeRequest) (dto.OtherFeeResponse, error)
	Update(ctx context.Context, request dto.UpdateOtherFeeRequest) (dto.OtherFeeResponse, error)
	Remove(ctx context.Context, groupExpenseID, otherFeeID, userProfileID uuid.UUID) error
	GetCalculationMethods(ctx context.Context) ([]dto.FeeCalculationMethodInfo, error)
}

type ExpenseBillService interface {
	Save(ctx context.Context, req *dto.NewExpenseBillRequest) (dto.ExpenseBillResponse, error)
	GetAllCreated(ctx context.Context, creatorProfileID uuid.UUID) ([]dto.ExpenseBillResponse, error)
	Get(ctx context.Context, profileID, id uuid.UUID) (dto.ExpenseBillResponse, error)
	Delete(ctx context.Context, profileID, id uuid.UUID) error
}
