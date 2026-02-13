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

type PlanService interface {
	Create(ctx context.Context, req dto.NewPlanRequest) (dto.PlanResponse, error)
	GetList(ctx context.Context) ([]dto.PlanResponse, error)
	GetOne(ctx context.Context, id uuid.UUID) (dto.PlanResponse, error)
	Update(ctx context.Context, req dto.UpdatePlanRequest) (dto.PlanResponse, error)
	Delete(ctx context.Context, id uuid.UUID) (dto.PlanResponse, error)
}

type planService struct {
	transactor      crud.Transactor
	planRepo        crud.Repository[entity.Plan]
	planVersionRepo crud.Repository[entity.PlanVersion]
}

func NewPlanService(
	transactor crud.Transactor,
	repo crud.Repository[entity.Plan],
	planVersionRepo crud.Repository[entity.PlanVersion],
) *planService {
	return &planService{
		transactor,
		repo,
		planVersionRepo,
	}
}

func (ps *planService) Create(ctx context.Context, req dto.NewPlanRequest) (dto.PlanResponse, error) {
	newPlan := entity.Plan{
		Name: req.Name,
	}

	insertedPlan, err := ps.planRepo.Insert(ctx, newPlan)
	if err != nil {
		return dto.PlanResponse{}, err
	}

	return mapper.PlanToResponse(insertedPlan), nil
}

func (ps *planService) GetList(ctx context.Context) ([]dto.PlanResponse, error) {
	spec := crud.Specification[entity.Plan]{}
	plans, err := ps.planRepo.FindAll(ctx, spec)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(plans, mapper.PlanToResponse), nil
}

func (ps *planService) GetOne(ctx context.Context, id uuid.UUID) (dto.PlanResponse, error) {
	plan, err := ps.getByID(ctx, id, false)
	if err != nil {
		return dto.PlanResponse{}, err
	}

	return mapper.PlanToResponse(plan), nil
}

func (ps *planService) Update(ctx context.Context, req dto.UpdatePlanRequest) (dto.PlanResponse, error) {
	var resp dto.PlanResponse
	err := ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		plan, err := ps.getByID(ctx, req.ID, true)
		if err != nil {
			return err
		}

		plan.Name = req.Name
		plan.IsActive = req.IsActive

		updatedPlan, err := ps.planRepo.Update(ctx, plan)
		if err != nil {
			return err
		}

		resp = mapper.PlanToResponse(updatedPlan)
		return nil
	})
	return resp, err
}

func (ps *planService) Delete(ctx context.Context, id uuid.UUID) (dto.PlanResponse, error) {
	var resp dto.PlanResponse
	err := ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		plan, err := ps.getByID(ctx, id, true)
		if err != nil {
			return err
		}

		versionSpec := crud.Specification[entity.PlanVersion]{}
		versionSpec.Model.PlanID = plan.ID
		planVersions, err := ps.planVersionRepo.FindAll(ctx, versionSpec)
		if err != nil {
			return err
		}
		if len(planVersions) > 0 {
			return ungerr.ConflictError("cannot delete plan, there exists plan versions for this plan")
		}

		if err = ps.planRepo.Delete(ctx, plan); err != nil {
			return err
		}

		resp = mapper.PlanToResponse(plan)
		return nil
	})
	return resp, err
}

func (ps *planService) getByID(ctx context.Context, id uuid.UUID, forUpdate bool) (entity.Plan, error) {
	spec := crud.Specification[entity.Plan]{}
	spec.Model.ID = id
	spec.ForUpdate = forUpdate
	plan, err := ps.planRepo.FindFirst(ctx, spec)
	if err != nil {
		return entity.Plan{}, err
	}
	if plan.IsZero() {
		return entity.Plan{}, ungerr.NotFoundError("plan is not found")
	}
	return plan, nil
}
