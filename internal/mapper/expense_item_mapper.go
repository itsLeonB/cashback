package mapper

import (
	expenseitemV1 "github.com/itsLeonB/billsplittr-protos/gen/go/expenseitem/v1"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/orcashtrator/internal/domain/expenseitem"

	"github.com/itsLeonB/orcashtrator/internal/dto"
)

func NewExpenseItemRequestToData(req dto.NewExpenseItemRequest) expenseitem.ExpenseItemData {
	return expenseitem.ExpenseItemData{
		Name:     req.Name,
		Amount:   req.Amount,
		Quantity: req.Quantity,
	}
}

func UpdateExpenseItemRequestToData(req dto.UpdateExpenseItemRequest) expenseitem.ExpenseItemData {
	return expenseitem.ExpenseItemData{
		Name:         req.Name,
		Amount:       req.Amount,
		Quantity:     req.Quantity,
		Participants: ezutil.MapSlice(req.Participants, itemParticipantRequestToData),
	}
}

func itemParticipantRequestToData(req dto.ItemParticipantRequest) expenseitem.ItemParticipant {
	return expenseitem.ItemParticipant{
		ProfileID: req.ProfileID,
		Share:     req.Share,
	}
}

func ToItemParticipantProto(req dto.ItemParticipantRequest) *expenseitemV1.ItemParticipant {
	return &expenseitemV1.ItemParticipant{
		ProfileId: req.ProfileID.String(),
		Share:     req.Share.InexactFloat64(),
	}
}
