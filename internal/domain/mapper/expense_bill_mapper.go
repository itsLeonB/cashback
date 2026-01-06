package mapper

import (
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
)

func ExpenseBillToResponse(
	bill expenses.ExpenseBill,
	url string,
) dto.ExpenseBillResponse {
	return dto.ExpenseBillResponse{
		BaseDTO:  BaseToDTO(bill.BaseEntity),
		ImageURL: url,
		Status:   bill.Status,
	}
}
