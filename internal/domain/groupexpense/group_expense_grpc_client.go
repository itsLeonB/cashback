package groupexpense

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/itsLeonB/billsplittr-protos/gen/go/groupexpense/v1"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/itsLeonB/orcashtrator/internal/domain/expenseitem"
	"github.com/itsLeonB/orcashtrator/internal/domain/otherfee"
	"github.com/rotisserie/eris"
	"golang.org/x/text/currency"
	"google.golang.org/grpc"
)

type GroupExpenseClient interface {
	CreateDraft(ctx context.Context, req CreateDraftRequest) (GroupExpense, error)
	GetAllCreated(ctx context.Context, profileID uuid.UUID, status appconstant.ExpenseStatus) ([]GroupExpense, error)
	GetDetails(ctx context.Context, id uuid.UUID) (GroupExpense, error)
	ConfirmDraft(ctx context.Context, req ConfirmDraftRequest) (GroupExpense, error)
}

type groupExpenseClient struct {
	validate *validator.Validate
	client   groupexpense.GroupExpenseServiceClient
}

func NewGroupExpenseClient(
	validate *validator.Validate,
	conn *grpc.ClientConn,
) GroupExpenseClient {
	if validate == nil {
		panic("validator is nil")
	}

	return &groupExpenseClient{
		validate,
		groupexpense.NewGroupExpenseServiceClient(conn),
	}
}

func (gec *groupExpenseClient) CreateDraft(ctx context.Context, req CreateDraftRequest) (GroupExpense, error) {
	if err := gec.validate.Struct(req); err != nil {
		return GroupExpense{}, err
	}

	fees, err := ezutil.MapSliceWithError(req.OtherFees, otherfee.ToOtherFeeProto)
	if err != nil {
		return GroupExpense{}, err
	}

	request := &groupexpense.CreateDraftRequest{
		CreatorProfileId: req.CreatorProfileID.String(),
		PayerProfileId:   req.PayerProfileID.String(),
		TotalAmount:      ezutil.DecimalToMoney(req.TotalAmount, currency.IDR.String()),
		Subtotal:         ezutil.DecimalToMoney(req.Subtotal, currency.IDR.String()),
		Description:      req.Description,
		Items:            ezutil.MapSlice(req.Items, expenseitem.ToExpenseItemProto),
		OtherFees:        fees,
	}

	response, err := gec.client.CreateDraft(ctx, request)
	if err != nil {
		return GroupExpense{}, err
	}

	return fromGroupExpenseProto(response.GetGroupExpense())
}

func (gec *groupExpenseClient) GetAllCreated(ctx context.Context, profileID uuid.UUID, status appconstant.ExpenseStatus) ([]GroupExpense, error) {
	expenseStatus, err := toExpenseStatusProto(status)
	if err != nil {
		return nil, err
	}

	request := &groupexpense.GetAllCreatedRequest{
		ProfileId: profileID.String(),
		Status:    expenseStatus,
	}

	response, err := gec.client.GetAllCreated(ctx, request)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSliceWithError(response.GetGroupExpenses(), fromGroupExpenseProto)
}

func (gec *groupExpenseClient) GetDetails(ctx context.Context, id uuid.UUID) (GroupExpense, error) {
	if id == uuid.Nil {
		return GroupExpense{}, eris.New("group expense id is nil")
	}

	request := &groupexpense.GetDetailsRequest{
		Id: id.String(),
	}

	response, err := gec.client.GetDetails(ctx, request)
	if err != nil {
		return GroupExpense{}, err
	}

	return fromGroupExpenseProto(response.GetGroupExpense())
}

func (gec *groupExpenseClient) ConfirmDraft(ctx context.Context, req ConfirmDraftRequest) (GroupExpense, error) {
	if err := gec.validate.Struct(req); err != nil {
		return GroupExpense{}, err
	}

	request := &groupexpense.ConfirmDraftRequest{
		Id:        req.ID.String(),
		ProfileId: req.ProfileID.String(),
		DryRun:    req.DryRun,
	}

	response, err := gec.client.ConfirmDraft(ctx, request)
	if err != nil {
		return GroupExpense{}, err
	}

	return fromGroupExpenseProto(response.GetGroupExpense())
}
