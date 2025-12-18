package mapper

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/orcashtrator/internal/domain/expensebill"
	"github.com/itsLeonB/orcashtrator/internal/dto"
)

func ExpenseBillToResponse(
	bill expensebill.ExpenseBill,
	url string,
	profileID uuid.UUID,
	namesByProfileIDs map[uuid.UUID]string,
) dto.ExpenseBillResponse {
	return dto.ExpenseBillResponse{
		ID:                 bill.ID,
		CreatorProfileID:   bill.CreatorProfileID,
		PayerProfileID:     bill.PayerProfileID,
		ImageURL:           url,
		Status:             bill.Status,
		CreatedAt:          bill.CreatedAt,
		UpdatedAt:          bill.UpdatedAt,
		DeletedAt:          bill.DeletedAt,
		IsCreatedByUser:    bill.CreatorProfileID == profileID,
		IsPaidByUser:       bill.PayerProfileID == profileID,
		CreatorProfileName: namesByProfileIDs[bill.CreatorProfileID],
		PayerProfileName:   namesByProfileIDs[bill.PayerProfileID],
	}
}

func ExpenseBillSimpleMapper(url string, profileID uuid.UUID, namesByProfileIDs map[uuid.UUID]string) func(expensebill.ExpenseBill) dto.ExpenseBillResponse {
	return func(eb expensebill.ExpenseBill) dto.ExpenseBillResponse {
		return ExpenseBillToResponse(eb, url, profileID, namesByProfileIDs)
	}
}

func UniqueBillProfileIDs(bills []expensebill.ExpenseBill) []uuid.UUID {
	idsMap := make(map[uuid.UUID]struct{})
	for _, bill := range bills {
		idsMap[bill.CreatorProfileID] = struct{}{}
		idsMap[bill.PayerProfileID] = struct{}{}
	}

	ids := make([]uuid.UUID, 0, len(idsMap))
	for id := range idsMap {
		ids = append(ids, id)
	}

	return ids
}
