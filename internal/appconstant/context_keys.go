package appconstant

type ctxKey string

const (
	ContextUserID          ctxKey = "userID"
	ContextProfileID       ctxKey = "profileID"
	ContextFriendshipID    ctxKey = "friendshipID"
	ContextGroupExpenseID  ctxKey = "groupExpenseID"
	ContextExpenseItemID   ctxKey = "expenseItemID"
	ContextOtherFeeID      ctxKey = "otherFeeID"
	ContextExpenseBillID   ctxKey = "expenseBillID"
	ContextProvider        ctxKey = "provider"
	ContextFriendRequestID ctxKey = "friendRequestID"
)

func (c ctxKey) String() string {
	return string(c)
}
