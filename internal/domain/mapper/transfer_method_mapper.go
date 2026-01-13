package mapper

import (
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
)

func TransferMethodToResponse(tm debts.TransferMethod, iconURL string) dto.TransferMethodResponse {
	return dto.TransferMethodResponse{
		BaseDTO:  BaseToDTO(tm.BaseEntity),
		Name:     tm.Name,
		Display:  tm.Display,
		IconURL:  iconURL,
		ParentID: tm.ParentID.UUID,
	}
}
