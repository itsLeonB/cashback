package monetization

import (
	"context"

	"github.com/google/uuid"
	dto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	mapper "github.com/itsLeonB/cashback/internal/domain/mapper/monetization"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type PlanVersionService interface {
	Create(ctx context.Context, req dto.NewPlanVersionRequest) (dto.PlanVersionResponse, error)
	GetList(ctx context.Context) ([]dto.PlanVersionResponse, error)
	GetOne(ctx context.Context, id uuid.UUID) (dto.PlanVersionResponse, error)
	Update(ctx context.Context, req dto.UpdatePlanVersionRequest) (dto.PlanVersionResponse, error)
	Delete(ctx context.Context, id uuid.UUID) (dto.PlanVersionResponse, error)
}

type planVersionService struct {
	transactor      crud.Transactor
	planVersionRepo crud.Repository[entity.PlanVersion]
	planRepo        crud.Repository[entity.Plan]
}

func NewPlanVersionService(
	transactor crud.Transactor,
	repo crud.Repository[entity.PlanVersion],
	planRepo crud.Repository[entity.Plan],
) *planVersionService {
	return &planVersionService{
		transactor,
		repo,
		planRepo,
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
		EffectiveTo:        req.EffectiveTo,
		IsDefault:          req.IsDefault,
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
		planVersion.EffectiveTo = req.EffectiveTo
		planVersion.IsDefault = req.IsDefault

		updatedPlanVersion, err := pvs.planVersionRepo.Update(ctx, planVersion)
		if err != nil {
			return err
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
