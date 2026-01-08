package mapper

import (
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/go-crud"
)

func BaseToDTO(be crud.BaseEntity) dto.BaseDTO {
	return dto.BaseDTO{
		ID:        be.ID,
		CreatedAt: be.CreatedAt,
		UpdatedAt: be.UpdatedAt,
	}
}
