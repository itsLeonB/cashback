package handler

import (
	"context"
	"fmt"
	"sync"

	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/domain/service"
)

func ExpenseConfirmedHandler(
	debtSvc service.DebtService,
	notificationSvc service.NotificationService,
) func(ctx context.Context, msg message.ExpenseConfirmed) error {
	return func(ctx context.Context, msg message.ExpenseConfirmed) error {
		var wg sync.WaitGroup
		errChan := make(chan error, 2)

		wg.Go(func() {
			if err := debtSvc.ProcessConfirmedGroupExpense(ctx, msg); err != nil {
				errChan <- err
			}
		})

		wg.Go(func() {
			if err := notificationSvc.HandleExpenseConfirmed(ctx, msg); err != nil {
				errChan <- err
			}
		})

		wg.Wait()
		close(errChan)

		var errs []error
		for err := range errChan {
			errs = append(errs, err)
		}

		if len(errs) > 0 {
			return fmt.Errorf("errors: %v", errs)
		}

		return nil
	}
}
