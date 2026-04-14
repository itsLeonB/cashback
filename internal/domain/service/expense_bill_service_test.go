package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/service/queue"
	"github.com/itsLeonB/cashback/internal/core/service/storage"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	config.Global = &config.Config{
		App: config.App{
			BucketNameExpenseBill: "test-bucket",
		},
	}
}

type mockBillRepo struct {
	mock.Mock
	crud.Repository[expenses.ExpenseBill]
}

func (m *mockBillRepo) Insert(ctx context.Context, bill expenses.ExpenseBill) (expenses.ExpenseBill, error) {
	args := m.Called(ctx, bill)
	return args.Get(0).(expenses.ExpenseBill), args.Error(1)
}

func (m *mockBillRepo) Update(ctx context.Context, bill expenses.ExpenseBill) (expenses.ExpenseBill, error) {
	args := m.Called(ctx, bill)
	return args.Get(0).(expenses.ExpenseBill), args.Error(1)
}

type mockTransactor struct {
	mock.Mock
}

func (m *mockTransactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (m *mockTransactor) Begin(ctx context.Context) (context.Context, error) { return ctx, nil }
func (m *mockTransactor) Commit(ctx context.Context) error                   { return nil }
func (m *mockTransactor) Rollback(ctx context.Context)                       {}

type mockGroupExpenseService struct {
	mock.Mock
	service.GroupExpenseService
}

func (m *mockBillRepo) CountUploadedByDateRange(ctx context.Context, profileID uuid.UUID, start, end time.Time) (int, error) {
	return 0, nil
}

func (m *mockGroupExpenseService) GetUnconfirmedForUpdate(ctx context.Context, profileID, id uuid.UUID) (expenses.GroupExpense, error) {
	args := m.Called(ctx, profileID, id)
	return args.Get(0).(expenses.GroupExpense), args.Error(1)
}

type mockSubscriptionLimitService struct {
	mock.Mock
	service.SubscriptionLimitService
}

func (m *mockSubscriptionLimitService) CheckUploadLimit(ctx context.Context, profileID uuid.UUID) error {
	args := m.Called(ctx, profileID)
	return args.Error(0)
}

type mockImageService struct {
	mock.Mock
	storage.ImageService
}

func (m *mockImageService) Upload(ctx context.Context, req *storage.ImageUploadRequest) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}

func (m *mockImageService) GetUploadURL(fileID storage.FileIdentifier) (string, error) {
	args := m.Called(fileID)
	return args.String(0), args.Error(1)
}

type mockTaskQueue struct {
	mock.Mock
}

func (m *mockTaskQueue) Enqueue(ctx context.Context, msg queue.TaskMessage) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *mockTaskQueue) AsyncEnqueue(ctx context.Context, msg queue.TaskMessage) {}
func (m *mockTaskQueue) Shutdown() error                                         { return nil }

func TestExpenseBillService_Save_ReuploadRules(t *testing.T) {
	ctx := context.Background()
	profileID := uuid.New()
	expenseID := uuid.New()
	billID := uuid.New()

	tests := []struct {
		name          string
		initialStatus expenses.BillStatus
		expectError   bool
		expectReuse   bool
		reuploadCount int
	}{
		{
			name:          "New Upload",
			initialStatus: "",
			expectError:   false,
			expectReuse:   false,
		},
		{
			name:          "Reuse NOT_UPLOADED",
			initialStatus: expenses.NotUploadedBill,
			expectError:   false,
			expectReuse:   true,
		},
		{
			name:          "Reuse FAILED_EXTRACTING",
			initialStatus: expenses.FailedExtracting,
			expectError:   false,
			expectReuse:   true,
		},
		{
			name:          "Reuse NOT_DETECTED",
			initialStatus: expenses.NotDetectedBill,
			expectError:   false,
			expectReuse:   true,
		},
		{
			name:          "Reject PENDING",
			initialStatus: expenses.PendingBill,
			expectError:   true,
		},
		{
			name:          "Reject EXTRACTED",
			initialStatus: expenses.ExtractedBill,
			expectError:   true,
		},
		{
			name:          "Reject PARSED",
			initialStatus: expenses.ParsedBill,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskQueue := new(mockTaskQueue)
			billRepo := new(mockBillRepo)
			transactor := new(mockTransactor)
			imageSvc := new(mockImageService)
			expenseSvc := new(mockGroupExpenseService)
			subsLimitSvc := new(mockSubscriptionLimitService)

			svc := service.NewExpenseBillService(taskQueue, billRepo, transactor, imageSvc, nil, expenseSvc, subsLimitSvc)

			subsLimitSvc.On("CheckUploadLimit", mock.Anything, profileID).Return(nil)

			expense := expenses.GroupExpense{
				BaseEntity: crud.BaseEntity{ID: expenseID},
			}
			if tt.initialStatus != "" {
				expense.Bill = expenses.ExpenseBill{
					BaseEntity:     crud.BaseEntity{ID: billID},
					GroupExpenseID: expenseID,
					Status:         tt.initialStatus,
					ImageName:      "existing_image.jpg",
				}
			}

			expenseSvc.On("GetUnconfirmedForUpdate", mock.Anything, profileID, expenseID).Return(expense, nil)

			req := &dto.NewExpenseBillRequest{
				ProfileID:      profileID,
				GroupExpenseID: expenseID,
				Filename:       "bill.jpg",
				ImageData:      []byte("fake image data"),
			}

			if tt.expectError {
				err := svc.Save(ctx, req)
				assert.Error(t, err)
				if appErr, ok := err.(ungerr.AppError); ok {
					assert.Equal(t, "Not allowed to reupload", appErr.Details())
				} else {
					t.Errorf("expected ungerr.AppError, got %T", err)
				}
				return
			}

			if tt.expectReuse {
				billRepo.On("Update", mock.Anything, mock.MatchedBy(func(b expenses.ExpenseBill) bool {
					return b.ID == billID && b.Status == expenses.PendingBill
				})).Return(expenses.ExpenseBill{BaseEntity: crud.BaseEntity{ID: billID}, ImageName: "existing_image.jpg"}, nil)
			} else {
				billRepo.On("Insert", mock.Anything, mock.MatchedBy(func(b expenses.ExpenseBill) bool {
					return b.GroupExpenseID == expenseID && b.Status == expenses.PendingBill
				})).Return(expenses.ExpenseBill{BaseEntity: crud.BaseEntity{ID: uuid.New()}, ImageName: "new_image.jpg"}, nil)
			}

			imageSvc.On("Upload", mock.Anything, mock.Anything).Return("uri", nil)
			taskQueue.On("Enqueue", mock.Anything, mock.Anything).Return(nil)

			err := svc.Save(ctx, req)
			assert.NoError(t, err)

			expenseSvc.AssertExpectations(t)
			billRepo.AssertExpectations(t)
		})
	}
}
