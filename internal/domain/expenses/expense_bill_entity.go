package expenses

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
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

type ExpenseBill struct {
	crud.BaseEntity
	GroupExpenseID uuid.UUID
	ImageName      string
	Status         BillStatus
	ExtractedText  string
}

func (eb ExpenseBill) TableName() string {
	return "group_expense_bills"
}
