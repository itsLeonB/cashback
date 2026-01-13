package mapper

import (
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
)

func ProfileTransferMethodPopulator(transferMethodPopulator func(debts.TransferMethod) dto.TransferMethodResponse) func(debts.ProfileTransferMethod) dto.ProfileTransferMethodResponse {
	return func(ptm debts.ProfileTransferMethod) dto.ProfileTransferMethodResponse {
		return dto.ProfileTransferMethodResponse{
			BaseDTO:       BaseToDTO(ptm.BaseEntity),
			Method:        transferMethodPopulator(ptm.Method),
			AccountName:   ptm.AccountName,
			AccountNumber: ptm.AccountNumber,
		}
	}
}
