package groupexpense

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/billsplittr-protos/gen/go/groupexpense/v1"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/itsLeonB/orcashtrator/internal/domain"
	"github.com/itsLeonB/orcashtrator/internal/domain/expenseitem"
	"github.com/itsLeonB/orcashtrator/internal/domain/otherfee"
	"github.com/rotisserie/eris"
)

func fromGroupExpenseProto(ge *groupexpense.GroupExpenseResponse) (GroupExpense, error) {
	if ge == nil {
		return GroupExpense{}, eris.New("group expense is nil")
	}

	creatorProfileID, err := ezutil.Parse[uuid.UUID](ge.GetCreatorProfileId())
	if err != nil {
		return GroupExpense{}, err
	}

	payerProfileID, err := ezutil.Parse[uuid.UUID](ge.GetPayerProfileId())
	if err != nil {
		return GroupExpense{}, err
	}

	items, err := ezutil.MapSliceWithError(ge.GetItems(), expenseitem.FromExpenseItemResponseProto)
	if err != nil {
		return GroupExpense{}, err
	}

	fees, err := ezutil.MapSliceWithError(ge.GetOtherFees(), otherfee.FromOtherFeeResponseProto)
	if err != nil {
		return GroupExpense{}, err
	}

	participants, err := ezutil.MapSliceWithError(ge.GetParticipants(), fromExpenseParticipantProto)
	if err != nil {
		return GroupExpense{}, err
	}

	metadata, err := domain.FromAuditMetadataProto(ge.GetAuditMetadata())
	if err != nil {
		return GroupExpense{}, err
	}

	status, err := fromExpenseStatusProto(ge.GetStatus())
	if err != nil {
		return GroupExpense{}, err
	}

	return GroupExpense{
		CreatorProfileID:        creatorProfileID,
		PayerProfileID:          payerProfileID,
		TotalAmount:             ezutil.MoneyToDecimal(ge.GetTotalAmount()),
		Subtotal:                ezutil.MoneyToDecimal(ge.GetSubtotal()),
		ItemsTotal:              ezutil.MoneyToDecimal(ge.GetItemsTotal()),
		FeesTotal:               ezutil.MoneyToDecimal(ge.GetFeesTotal()),
		Description:             ge.GetDescription(),
		IsConfirmed:             ge.GetIsConfirmed(),
		IsParticipantsConfirmed: ge.GetIsParticipantsConfirmed(),
		Status:                  status,
		Items:                   items,
		OtherFees:               fees,
		Participants:            participants,
		AuditMetadata:           metadata,
	}, nil
}

func fromExpenseStatusProto(status groupexpense.GroupExpenseResponse_ExpenseStatus) (appconstant.ExpenseStatus, error) {
	switch status {
	case groupexpense.GroupExpenseResponse_EXPENSE_STATUS_UNSPECIFIED:
		return "", eris.New("unspecified expense status enum")
	case groupexpense.GroupExpenseResponse_EXPENSE_STATUS_DRAFT:
		return appconstant.DraftExpense, nil
	case groupexpense.GroupExpenseResponse_EXPENSE_STATUS_PROCESSING_BILL:
		return appconstant.ProcessingBillExpense, nil
	case groupexpense.GroupExpenseResponse_EXPENSE_STATUS_READY:
		return appconstant.ReadyExpense, nil
	case groupexpense.GroupExpenseResponse_EXPENSE_STATUS_CONFIRMED:
		return appconstant.ConfirmedExpense, nil
	default:
		return "", eris.Errorf("unknown expense status enum: %s", status.String())
	}
}

func fromExpenseParticipantProto(ep *groupexpense.ExpenseParticipantResponse) (ExpenseParticipant, error) {
	if ep == nil {
		return ExpenseParticipant{}, eris.New("expense participant response is nil")
	}

	profileID, err := ezutil.Parse[uuid.UUID](ep.GetProfileId())
	if err != nil {
		return ExpenseParticipant{}, err
	}

	return ExpenseParticipant{
		ProfileID:   profileID,
		ShareAmount: ezutil.MoneyToDecimal(ep.GetShareAmount()),
	}, nil
}
