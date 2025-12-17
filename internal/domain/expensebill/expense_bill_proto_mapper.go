package expensebill

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/billsplittr-protos/gen/go/expensebill/v1"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/orcashtrator/internal/domain"
	"github.com/rotisserie/eris"
)

func toExpenseBillProto(bill ExpenseBill) *expensebill.ExpenseBill {
	return &expensebill.ExpenseBill{
		CreatorProfileId: bill.CreatorProfileID.String(),
		PayerProfileId:   bill.PayerProfileID.String(),
		GroupExpenseId:   bill.GroupExpenseID.String(),
		ObjectKey:        bill.ObjectKey,
	}
}

func fromExpenseBillProto(bill *expensebill.ExpenseBillResponse) (ExpenseBill, error) {
	if bill == nil {
		return ExpenseBill{}, eris.New("expense bill response is nil")
	}

	data := bill.GetExpenseBill()
	if data == nil {
		return ExpenseBill{}, eris.New("expense bill is nil")
	}

	creatorProfileID, err := ezutil.Parse[uuid.UUID](data.GetCreatorProfileId())
	if err != nil {
		return ExpenseBill{}, err
	}

	payerProfileID, err := ezutil.Parse[uuid.UUID](data.GetPayerProfileId())
	if err != nil {
		return ExpenseBill{}, err
	}

	metadata, err := domain.FromAuditMetadataProto(bill.GetAuditMetadata())
	if err != nil {
		return ExpenseBill{}, err
	}

	return ExpenseBill{
		CreatorProfileID: creatorProfileID,
		PayerProfileID:   payerProfileID,
		ObjectKey:        data.GetObjectKey(),
		AuditMetadata:    metadata,
	}, nil
}
