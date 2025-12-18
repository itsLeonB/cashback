package appconstant

type DebtTransactionType string
type FriendshipType string
type FeeCalculationMethod string

const (
	Lend  DebtTransactionType = "LEND"
	Repay DebtTransactionType = "REPAY"

	Real      FriendshipType = "REAL"
	Anonymous FriendshipType = "ANON"

	GroupExpenseTransferMethod = "GROUP_EXPENSE"

	EqualSplitFee    FeeCalculationMethod = "EQUAL_SPLIT"
	ItemizedSplitFee FeeCalculationMethod = "ITEMIZED_SPLIT"
)

type ExpenseStatus string

const (
	DraftExpense     ExpenseStatus = "DRAFT"
	ReadyExpense     ExpenseStatus = "READY"
	ConfirmedExpense ExpenseStatus = "CONFIRMED"
)

type BillStatus string

const (
	PendingBill BillStatus = "PENDING"
	ParsedBill  BillStatus = "PARSED"
	FailedBill  BillStatus = "FAILED"
)
