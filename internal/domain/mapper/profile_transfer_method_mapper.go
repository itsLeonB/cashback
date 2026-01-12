package mapper

import (
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
)

func ProfileTransferMethodToResponse(ptm debts.ProfileTransferMethod) dto.ProfileTransferMethodResponse {
	return dto.ProfileTransferMethodResponse{
		BaseDTO:       BaseToDTO(ptm.BaseEntity),
		Method:        TransferMethodToResponse(ptm.Method),
		AccountName:   ptm.AccountName,
		AccountNumber: ptm.AccountNumber,
	}
}
