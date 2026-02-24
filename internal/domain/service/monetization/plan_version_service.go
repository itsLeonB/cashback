package monetization

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	dto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	mapper "github.com/itsLeonB/cashback/internal/domain/mapper/monetization"
	"github.com/itsLeonB/cashback/internal/domain/repository/monetization"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type PlanVersionService interface {
	// Admin
	Create(ctx context.Context, req dto.NewPlanVersionRequest) (dto.PlanVersionResponse, error)
	GetList(ctx context.Context) ([]dto.PlanVersionResponse, error)
	GetOne(ctx context.Context, id uuid.UUID) (dto.PlanVersionResponse, error)
	Update(ctx context.Context, req dto.UpdatePlanVersionRequest) (dto.PlanVersionResponse, error)
	Delete(ctx context.Context, id uuid.UUID) (dto.PlanVersionResponse, error)

	// Public
	GetActive(ctx context.Context) ([]dto.PlanVersionResponse, error)
}

type planVersionService struct {
	transactor      crud.Transactor
	planVersionRepo monetization.PlanVersionRepository
}

func NewPlanVersionService(
	transactor crud.Transactor,
	repo monetization.PlanVersionRepository,
) *planVersionService {
	return &planVersionService{
		transactor,
		repo,
	}
}

func (pvs *planVersionService) Create(ctx context.Context, req dto.NewPlanVersionRequest) (dto.PlanVersionResponse, error) {
	newPlanVersion := entity.PlanVersion{
		PlanID:             req.PlanID,
		PriceAmount:        req.PriceAmount,
		PriceCurrency:      req.PriceCurrency,
		BillingInterval:    entity.BillingInterval(req.BillingInterval),
		BillUploadsDaily:   req.BillUploadsDaily,
		BillUploadsMonthly: req.BillUploadsMonthly,
		EffectiveFrom:      req.EffectiveFrom,
		IsDefault:          req.IsDefault,
	}

	if !req.EffectiveTo.IsZero() {
		newPlanVersion.EffectiveTo = sql.NullTime{
			Time:  req.EffectiveTo,
			Valid: true,
		}
	}

	insertedPlanVersion, err := pvs.planVersionRepo.Insert(ctx, newPlanVersion)
	if err != nil {
		return dto.PlanVersionResponse{}, err
	}

	return mapper.PlanVersionToResponse(insertedPlanVersion), nil
}

func (pvs *planVersionService) GetList(ctx context.Context) ([]dto.PlanVersionResponse, error) {
	spec := crud.Specification[entity.PlanVersion]{}
	spec.PreloadRelations = []string{"Plan"}
	planVersions, err := pvs.planVersionRepo.FindAll(ctx, spec)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(planVersions, mapper.PlanVersionToResponse), nil
}

func (pvs *planVersionService) GetOne(ctx context.Context, id uuid.UUID) (dto.PlanVersionResponse, error) {
	planVersion, err := pvs.getByID(ctx, id, false, []string{"Plan"})
	if err != nil {
		return dto.PlanVersionResponse{}, err
	}

	return mapper.PlanVersionToResponse(planVersion), nil
}

func (pvs *planVersionService) Update(ctx context.Context, req dto.UpdatePlanVersionRequest) (dto.PlanVersionResponse, error) {
	var resp dto.PlanVersionResponse
	err := pvs.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		planVersion, err := pvs.getByID(ctx, req.ID, true, nil)
		if err != nil {
			return err
		}

		planVersion.PlanID = req.PlanID
		planVersion.PriceAmount = req.PriceAmount
		planVersion.PriceCurrency = req.PriceCurrency
		planVersion.BillingInterval = entity.BillingInterval(req.BillingInterval)
		planVersion.BillUploadsDaily = req.BillUploadsDaily
		planVersion.BillUploadsMonthly = req.BillUploadsMonthly
		planVersion.EffectiveFrom = req.EffectiveFrom
		planVersion.IsDefault = req.IsDefault

		planVersion.EffectiveTo = sql.NullTime{
			Time:  req.EffectiveTo,
			Valid: !req.EffectiveTo.IsZero(),
		}

		updatedPlanVersion, err := pvs.planVersionRepo.Update(ctx, planVersion)
		if err != nil {
			return err
		}

		if req.IsDefault {
			if err = pvs.planVersionRepo.SetAsDefault(ctx, updatedPlanVersion.ID); err != nil {
				return err
			}
		}

		resp = mapper.PlanVersionToResponse(updatedPlanVersion)
		return nil
	})
	return resp, err
}

func (pvs *planVersionService) Delete(ctx context.Context, id uuid.UUID) (dto.PlanVersionResponse, error) {
	var resp dto.PlanVersionResponse
	err := pvs.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		planVersion, err := pvs.getByID(ctx, id, true, nil)
		if err != nil {
			return err
		}

		if err = pvs.planVersionRepo.Delete(ctx, planVersion); err != nil {
			return err
		}

		resp = mapper.PlanVersionToResponse(planVersion)
		return nil
	})
	return resp, err
}

func (pvs *planVersionService) GetActive(ctx context.Context) ([]dto.PlanVersionResponse, error) {
	spec := crud.Specification[entity.PlanVersion]{}
	spec.PreloadRelations = []string{"Plan"}
	planVersions, err := pvs.planVersionRepo.FindAll(ctx, spec)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	versionsByPlanID := make(map[uuid.UUID][]entity.PlanVersion)
	for _, planVersion := range planVersions {
		if planVersion.EffectiveFrom.After(now) || (planVersion.EffectiveTo.Valid && planVersion.EffectiveTo.Time.Before(now)) || !planVersion.Plan.IsActive {
			continue
		}
		if _, exists := versionsByPlanID[planVersion.PlanID]; !exists {
			versionsByPlanID[planVersion.PlanID] = []entity.PlanVersion{}
		}
		versionsByPlanID[planVersion.PlanID] = append(versionsByPlanID[planVersion.PlanID], planVersion)
	}

	responses := make([]dto.PlanVersionResponse, len(versionsByPlanID))
	for _, versions := range versionsByPlanID {
		responses = append(responses, mapper.PlanVersionToResponse(versions[0]))
	}

	return responses, nil
}

func (pvs *planVersionService) getByID(ctx context.Context, id uuid.UUID, forUpdate bool, relations []string) (entity.PlanVersion, error) {
	spec := crud.Specification[entity.PlanVersion]{}
	spec.Model.ID = id
	spec.ForUpdate = forUpdate
	spec.PreloadRelations = relations
	planVersion, err := pvs.planVersionRepo.FindFirst(ctx, spec)
	if err != nil {
		return entity.PlanVersion{}, err
	}
	if planVersion.IsZero() {
		return entity.PlanVersion{}, ungerr.NotFoundError("plan version is not found")
	}
	return planVersion, nil
}
