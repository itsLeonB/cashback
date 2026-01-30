package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/cashback/internal/domain/message"
)

type UserService interface {
	CreateNew(ctx context.Context, request dto.NewUserRequest) (users.User, error)
	FindByEmail(ctx context.Context, email string) (users.User, error)
	Verify(ctx context.Context, id uuid.UUID, email string, name string, avatar string) (users.User, error)
	GeneratePasswordResetToken(ctx context.Context, userID uuid.UUID) (string, error)
	ResetPassword(ctx context.Context, userID uuid.UUID, email, resetToken, password string) (users.User, error)

	GetByID(ctx context.Context, id uuid.UUID) (users.User, error)
}

type AuthService interface {
	Register(ctx context.Context, request dto.RegisterRequest) (dto.RegisterResponse, error)
	InternalLogin(ctx context.Context, request dto.InternalLoginRequest) (dto.LoginResponse, error)
	VerifyToken(ctx context.Context, token string) (bool, map[string]any, error)
	GetOAuth2URL(ctx context.Context, provider string) (string, error)
	OAuth2Login(ctx context.Context, provider, code, state string) (dto.LoginResponse, error)
	VerifyRegistration(ctx context.Context, token string) (dto.LoginResponse, error)
	SendPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) (dto.LoginResponse, error)
	RefreshToken(ctx context.Context, request dto.RefreshTokenRequest) (dto.RefreshTokenResponse, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
}

type OAuthService interface {
	GetOAuthURL(ctx context.Context, provider string) (string, error)
	HandleOAuthCallback(ctx context.Context, data dto.OAuthCallbackData) (dto.LoginResponse, error)
	CreateLoginResponse(user users.User, session users.Session) (dto.LoginResponse, error)
}

type ProfileService interface {
	Create(ctx context.Context, request dto.NewProfileRequest) (dto.ProfileResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (dto.ProfileResponse, error)
	Update(ctx context.Context, id uuid.UUID, name string) (dto.ProfileResponse, error)
	Search(ctx context.Context, profileID uuid.UUID, input string) ([]dto.ProfileResponse, error)
	Associate(ctx context.Context, userProfileID, realProfileID, anonProfileID uuid.UUID) error
	GetByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]dto.ProfileResponse, error)
	GetRealProfileID(ctx context.Context, anonProfileID uuid.UUID) (uuid.UUID, error)
	GetEntityByID(ctx context.Context, id uuid.UUID) (users.UserProfile, error)
	GetAssociatedIDs(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error)
}

type FriendshipService interface {
	CreateAnonymous(ctx context.Context, request dto.NewAnonymousFriendshipRequest) (dto.FriendshipResponse, error)
	GetAll(ctx context.Context, profileID uuid.UUID) ([]dto.FriendshipResponse, error)
	GetDetails(ctx context.Context, profileID, friendshipID uuid.UUID) (dto.FriendDetails, error)
	IsFriends(ctx context.Context, profileID1, profileID2 uuid.UUID) (bool, bool, error)
	CreateReal(ctx context.Context, userProfileID, friendProfileID uuid.UUID) (dto.FriendshipResponse, error)
	GetByProfileIDs(ctx context.Context, profileID1, profileID2 uuid.UUID) (users.Friendship, error)

	ConstructNotification(ctx context.Context, msg message.FriendRequestAccepted) (entity.Notification, error)
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

	ConstructNotification(ctx context.Context, msg message.FriendRequestSent) (entity.Notification, error)
}

type FriendDetailsService interface {
	GetDetails(ctx context.Context, profileID, friendshipID uuid.UUID) (dto.FriendDetailsResponse, error)
}

type DebtService interface {
	RecordNewTransaction(ctx context.Context, request dto.NewDebtTransactionRequest) (dto.DebtTransactionResponse, error)
	GetTransactions(ctx context.Context, userProfileID uuid.UUID) ([]dto.DebtTransactionResponse, error)
	GetAllByProfileIDs(ctx context.Context, userProfileID, friendProfileID uuid.UUID) ([]debts.DebtTransaction, []uuid.UUID, error)
	GetTransactionSummary(ctx context.Context, profileID uuid.UUID) (dto.FriendBalance, error)
	GetRecent(ctx context.Context, profileID uuid.UUID) ([]dto.DebtTransactionResponse, error)

	ConstructNotification(ctx context.Context, msg message.DebtCreated) (entity.Notification, error)
	ProcessConfirmedGroupExpense(ctx context.Context, groupExpense expenses.GroupExpense) error
}

type TransferMethodService interface {
	GetAll(ctx context.Context, filter debts.ParentFilter, profileID uuid.UUID) ([]dto.TransferMethodResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (debts.TransferMethod, error)
	GetByName(ctx context.Context, name string) (debts.TransferMethod, error)
	SyncMethods(ctx context.Context) error
	SignedURLPopulator(ctx context.Context) func(debts.TransferMethod) dto.TransferMethodResponse
	Shutdown() error
}

type GroupExpenseService interface {
	CreateDraft(ctx context.Context, userProfileID uuid.UUID, description string) (dto.GroupExpenseResponse, error)
	GetAll(ctx context.Context, userProfileID uuid.UUID, ownership expenses.ExpenseOwnership, status expenses.ExpenseStatus) ([]dto.GroupExpenseResponse, error)
	GetDetails(ctx context.Context, id, userProfileID uuid.UUID) (dto.GroupExpenseResponse, error)
	ConfirmDraft(ctx context.Context, id, userProfileID uuid.UUID, dryRun bool) (dto.ExpenseConfirmationResponse, error)
	Delete(ctx context.Context, userProfileID, id uuid.UUID) error
	SyncParticipants(ctx context.Context, req dto.ExpenseParticipantsRequest) error
	GetRecent(ctx context.Context, profileID uuid.UUID) ([]dto.GroupExpenseResponse, error)

	GetUnconfirmedGroupExpenseForUpdate(ctx context.Context, profileID, id uuid.UUID) (expenses.GroupExpense, error)
	ParseFromBillText(ctx context.Context, msg message.ExpenseBillTextExtracted) error
	Recalculate(ctx context.Context, userProfileID, groupExpenseID uuid.UUID, amountChanged bool) error
	GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (expenses.GroupExpense, error)
	ConstructNotifications(ctx context.Context, msg message.ExpenseConfirmed) ([]entity.Notification, error)
	ProcessCallback(ctx context.Context, id uuid.UUID, callbackFn func(context.Context, expenses.GroupExpense) error) error
}

type ExpenseItemService interface {
	Add(ctx context.Context, request dto.NewExpenseItemRequest) error
	Update(ctx context.Context, request dto.UpdateExpenseItemRequest) error
	Remove(ctx context.Context, groupExpenseID, expenseItemID, userProfileID uuid.UUID) error
	SyncParticipants(ctx context.Context, req dto.SyncItemParticipantsRequest) error
}

type OtherFeeService interface {
	Add(ctx context.Context, request dto.NewOtherFeeRequest) (dto.OtherFeeResponse, error)
	Update(ctx context.Context, request dto.UpdateOtherFeeRequest) (dto.OtherFeeResponse, error)
	Remove(ctx context.Context, groupExpenseID, otherFeeID, userProfileID uuid.UUID) error
	GetCalculationMethods(ctx context.Context) []dto.FeeCalculationMethodInfo
}

type ExpenseBillService interface {
	Save(ctx context.Context, req *dto.NewExpenseBillRequest) error
	ExtractBillText(ctx context.Context, msg message.ExpenseBillUploaded) error
	Cleanup(ctx context.Context) error
	TriggerParsing(ctx context.Context, expenseID, billID uuid.UUID) error
}

type ProfileTransferMethodService interface {
	Add(ctx context.Context, req dto.NewProfileTransferMethodRequest) error
	GetAllByProfileID(ctx context.Context, profileID uuid.UUID) ([]dto.ProfileTransferMethodResponse, error)
	GetAllByFriendProfileID(ctx context.Context, userProfileID, friendProfileID uuid.UUID) ([]dto.ProfileTransferMethodResponse, error)
}

type NotificationService interface {
	HandleDebtCreated(ctx context.Context, msg message.DebtCreated) error
	HandleFriendRequestSent(ctx context.Context, msg message.FriendRequestSent) error
	HandleFriendRequestAccepted(ctx context.Context, msg message.FriendRequestAccepted) error
	HandleExpenseConfirmed(ctx context.Context, msg message.ExpenseConfirmed) error

	GetUnread(ctx context.Context, profileID uuid.UUID) ([]dto.NotificationResponse, error)
	MarkAsRead(ctx context.Context, profileID, notificationID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, profileID uuid.UUID) error
}

type PushNotificationService interface {
	Subscribe(ctx context.Context, req dto.PushSubscriptionRequest) error
	Unsubscribe(ctx context.Context, req dto.PushUnsubscribeRequest) error
	UnsubscribeBySession(ctx context.Context, sessionID uuid.UUID) error
	Deliver(ctx context.Context, msg message.NotificationCreated) error
}
