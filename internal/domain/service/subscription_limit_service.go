package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	"github.com/itsLeonB/cashback/internal/domain/service/monetization"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/ungerr"
)

type subscriptionLimitService struct {
	subscriptionSvc monetization.SubscriptionService
	billRepo        repository.ExpenseBillRepository
}

func NewSubscriptionLimitService(
	subscriptionSvc monetization.SubscriptionService,
	billRepo repository.ExpenseBillRepository,
) *subscriptionLimitService {
	return &subscriptionLimitService{
		subscriptionSvc,
		billRepo,
	}
}

func (sls *subscriptionLimitService) GetCurrent(ctx context.Context, profileID uuid.UUID) (dto.SubscriptionResponse, error) {
	currentSubs, err := sls.subscriptionSvc.GetCurrentSubscription(ctx, profileID, true)
	if err != nil {
		return dto.SubscriptionResponse{}, err
	}

	now := time.Now()
	year, month, day := now.Date()

	dailyLimits, err := sls.getDailyUploadLimit(ctx, int(currentSubs.PlanVersion.BillUploadsDaily), profileID, year, int(month), day)
	if err != nil {
		return dto.SubscriptionResponse{}, err
	}

	monthlyLimits, err := sls.getMonthlyUploadLimit(ctx, int(currentSubs.PlanVersion.BillUploadsMonthly), profileID, year, month)
	if err != nil {
		return dto.SubscriptionResponse{}, err
	}

	return dto.SubscriptionResponse{
		Plan: currentSubs.PlanVersion.Plan.Name,
		Limits: dto.Limits{
			Uploads: dto.UploadLimits{
				Daily:     dailyLimits,
				Monthly:   monthlyLimits,
				CanUpload: dailyLimits.CanUpload && monthlyLimits.CanUpload,
			},
		},
	}, nil
}

func (sls *subscriptionLimitService) CheckUploadLimit(ctx context.Context, profileID uuid.UUID) error {
	subscription, err := sls.subscriptionSvc.GetCurrentSubscription(ctx, profileID, true)
	if err != nil {
		return err
	}

	// No limits set - early return
	if subscription.PlanVersion.BillUploadsDaily == 0 && subscription.PlanVersion.BillUploadsMonthly == 0 {
		return nil
	}

	now := time.Now()
	year, month, day := now.Date()

	// Check daily limit first (more restrictive, faster to fail)
	if subscription.PlanVersion.BillUploadsDaily > 0 {
		uploadLimit, err := sls.getDailyUploadLimit(ctx, int(subscription.PlanVersion.BillUploadsDaily), profileID, year, int(month), day)
		if err != nil {
			return err
		}

		if !uploadLimit.CanUpload {
			return ungerr.ForbiddenError("bill uploads for today has reached current plan limit")
		}
	}

	// Only check monthly if daily passed
	if subscription.PlanVersion.BillUploadsMonthly > 0 {
		uploadLimit, err := sls.getMonthlyUploadLimit(ctx, int(subscription.PlanVersion.BillUploadsMonthly), profileID, year, month)
		if err != nil {
			return err
		}

		if !uploadLimit.CanUpload {
			return ungerr.ForbiddenError("bill uploads for this month has reached current plan limit")
		}
	}

	return nil
}

func (sls *subscriptionLimitService) getDailyUploadLimit(
	ctx context.Context,
	dailyLimit int,
	profileID uuid.UUID,
	year int,
	month int,
	day int,
) (dto.UploadLimit, error) {
	startOfDay, err := ezutil.GetStartOfDay(year, month, day)
	if err != nil {
		return dto.UploadLimit{}, err
	}

	endOfDay, err := ezutil.GetEndOfDay(year, month, day)
	if err != nil {
		return dto.UploadLimit{}, err
	}

	dailyCount, err := sls.billRepo.CountUploadedByDateRange(ctx, profileID, startOfDay, endOfDay)
	if err != nil {
		return dto.UploadLimit{}, err
	}

	return getUploadLimit(dailyLimit, dailyCount, startOfDay.AddDate(0, 0, 1)), nil
}

func (sls *subscriptionLimitService) getMonthlyUploadLimit(
	ctx context.Context,
	monthlyLimit int,
	profileID uuid.UUID,
	year int,
	month time.Month,
) (dto.UploadLimit, error) {
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	nextMonth := startOfMonth.AddDate(0, 1, 0)
	endOfMonth := nextMonth.Add(-time.Nanosecond)

	monthlyCount, err := sls.billRepo.CountUploadedByDateRange(ctx, profileID, startOfMonth, endOfMonth)
	if err != nil {
		return dto.UploadLimit{}, err
	}

	return getUploadLimit(monthlyLimit, monthlyCount, nextMonth), nil
}

func getUploadLimit(limit, count int, resetAt time.Time) dto.UploadLimit {
	remaining := max(limit-count, 0)
	canUpload := canUpload(limit, remaining)

	return dto.UploadLimit{
		Used:      count,
		Limit:     limit,
		Remaining: remaining,
		ResetAt:   resetAt,
		CanUpload: canUpload,
	}
}

func canUpload(limit, remaining int) bool {
	if limit == 0 {
		return true
	}

	return remaining > 0
}
