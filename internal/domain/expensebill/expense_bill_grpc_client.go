package expensebill

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/itsLeonB/billsplittr-protos/gen/go/expensebill/v1"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/rotisserie/eris"
	"google.golang.org/grpc"
)

type ExpenseBillClient interface {
	Save(ctx context.Context, req ExpenseBill) (ExpenseBill, error)
	GetAllCreated(ctx context.Context, creatorProfileID uuid.UUID) ([]ExpenseBill, error)
	Get(ctx context.Context, id uuid.UUID) (ExpenseBill, error)
	Delete(ctx context.Context, profileID, id uuid.UUID) error
}

type expenseBillClient struct {
	validate *validator.Validate
	client   expensebill.ExpenseBillServiceClient
}

func NewExpenseBillClient(
	validate *validator.Validate,
	conn *grpc.ClientConn,
) ExpenseBillClient {
	if validate == nil {
		panic("validator is nil")
	}

	return &expenseBillClient{
		validate,
		expensebill.NewExpenseBillServiceClient(conn),
	}
}

func (ebc *expenseBillClient) Save(ctx context.Context, req ExpenseBill) (ExpenseBill, error) {
	if err := ebc.validate.Struct(req); err != nil {
		return ExpenseBill{}, eris.Wrap(err, appconstant.ErrStructValidation)
	}

	request := expensebill.SaveRequest{
		ExpenseBill: toExpenseBillProto(req),
	}

	response, err := ebc.client.Save(ctx, &request)
	if err != nil {
		return ExpenseBill{}, err
	}

	return FromExpenseBillProto(response.GetExpenseBill())
}

func (ebc *expenseBillClient) GetAllCreated(ctx context.Context, creatorProfileID uuid.UUID) ([]ExpenseBill, error) {
	request := expensebill.GetAllCreatedRequest{
		CreatorProfileId: creatorProfileID.String(),
	}

	response, err := ebc.client.GetAllCreated(ctx, &request)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSliceWithError(response.GetExpenseBills(), FromExpenseBillProto)
}

func (ebc *expenseBillClient) Get(ctx context.Context, id uuid.UUID) (ExpenseBill, error) {
	request := expensebill.GetRequest{
		Id: id.String(),
	}

	response, err := ebc.client.Get(ctx, &request)
	if err != nil {
		return ExpenseBill{}, err
	}

	return FromExpenseBillProto(response.GetExpenseBill())
}

func (ebc *expenseBillClient) Delete(ctx context.Context, profileID, id uuid.UUID) error {
	request := expensebill.DeleteRequest{
		ProfileId: profileID.String(),
		Id:        id.String(),
	}

	_, err := ebc.client.Delete(ctx, &request)

	return err
}
