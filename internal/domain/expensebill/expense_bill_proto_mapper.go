package expensebill

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/billsplittr-protos/gen/go/expensebill/v1"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
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

func FromExpenseBillProto(bill *expensebill.ExpenseBillResponse) (ExpenseBill, error) {
	if bill == nil {
		return ExpenseBill{}, nil
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

	status, err := fromBillStatusProto(data.GetStatus())
	if err != nil {
		return ExpenseBill{}, err
	}

	return ExpenseBill{
		CreatorProfileID: creatorProfileID,
		PayerProfileID:   payerProfileID,
		ObjectKey:        data.GetObjectKey(),
		AuditMetadata:    metadata,
		Status:           status,
	}, nil
}

func fromBillStatusProto(status expensebill.ExpenseBill_Status) (appconstant.BillStatus, error) {
	switch status {
	case expensebill.ExpenseBill_STATUS_UNSPECIFIED:
		return "", eris.New("unspecified expense bill status enum")
	case expensebill.ExpenseBill_STATUS_PENDING:
		return appconstant.PendingBill, nil
	case expensebill.ExpenseBill_STATUS_EXTRACTED:
		return appconstant.ExtractedBill, nil
	case expensebill.ExpenseBill_STATUS_FAILED_EXTRACTING:
		return appconstant.FailedExtracting, nil
	case expensebill.ExpenseBill_STATUS_PARSED:
		return appconstant.ParsedBill, nil
	case expensebill.ExpenseBill_STATUS_FAILED_PARSING:
		return appconstant.FailedParsingBill, nil
	case expensebill.ExpenseBill_STATUS_NOT_DETECTED:
		return appconstant.NotDetectedBill, nil
	default:
		return "", eris.Errorf("unknown expense bill status enum: %s", status.String())
	}
}
