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
	PendingBill       BillStatus = "PENDING"           // Newly uploaded bill, to be extracted by OCR service
	ExtractedBill     BillStatus = "EXTRACTED"         // Text extracted from image, to be parsed by AI service
	FailedExtracting  BillStatus = "FAILED_EXTRACTING" // OCR failed to extract from image, retryable
	ParsedBill        BillStatus = "PARSED"            // Expense data parsed from text, considered finished, cannot retry
	FailedParsingBill BillStatus = "FAILED_PARSING"    // AI failed to parse from text, retryable
	NotDetectedBill   BillStatus = "NOT_DETECTED"      // AI cannot detect data from text, can upload another image
)
