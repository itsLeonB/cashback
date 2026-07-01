package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/cashback/internal/mocks"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestGroupExpenseService(
	t *testing.T,
) (service.GroupExpenseService, *mocks.MockRepository[expenses.ExpenseBill], *mocks.MockTransactor, *mocks.MockClient) {
	billRepo := mocks.NewMockRepository[expenses.ExpenseBill](t)
	transactor := mocks.NewMockTransactor(t)
	langfuseClient := mocks.NewMockClient(t)

	svc := service.NewGroupExpenseService(
		nil, // friendshipService (unused by ParseFromBillText)
		nil, // expenseRepo (unused: LLM call fails before this path is reached)
		transactor,
		nil, // feeCalculatorRegistry (unused by ParseFromBillText)
		nil, // otherFeeRepository (unused by ParseFromBillText)
		billRepo,
		nil, // llmService (unused: langfuseClient.GetPrompt fails first)
		nil, // imageSvc (unused by ParseFromBillText)
		nil, // taskQueue (unused by ParseFromBillText)
		langfuseClient,
		nil, // profileSvc (unused by ParseFromBillText)
	)

	return svc, billRepo, transactor, langfuseClient
}

func TestParseFromBillText_CallsLLMBeforeOpeningTransaction(t *testing.T) {
	svc, billRepo, transactor, langfuseClient := newTestGroupExpenseService(t)

	billID := uuid.New()
	pendingBill := expenses.ExpenseBill{
		BaseEntity:    crud.BaseEntity{ID: billID},
		Status:        expenses.ExtractedBill,
		ExtractedText: "total: 10.00",
	}

	// MockRepository[T] (internal/mocks/mock_crud_repository.go) is a hand-written
	// generic mock without the mockery "_Expecter" pattern, so it's driven via the
	// plain testify .On(...) API (see existing usage in profile_service_test.go),
	// unlike langfuseClient/transactor below which are mockery-generated with .EXPECT().
	billRepo.On("FindFirst", mock.Anything, mock.Anything).Return(pendingBill, nil)

	getPromptCall := langfuseClient.EXPECT().
		GetPrompt(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, ungerr.Unknown("langfuse unavailable"))

	withinTxCall := transactor.EXPECT().
		WithinTransaction(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	billRepo.On("Update", mock.Anything, mock.MatchedBy(func(b expenses.ExpenseBill) bool {
		return b.Status == expenses.FailedParsingBill
	})).Return(expenses.ExpenseBill{}, nil)

	mock.InOrder(getPromptCall.Call, withinTxCall.Call)

	err := svc.ParseFromBillText(context.Background(), message.ExpenseBillTextExtracted{ID: billID})

	assert.NoError(t, err)
}
