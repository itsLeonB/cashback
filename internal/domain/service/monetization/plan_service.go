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
	Delete(ctx context.Context, id uuid.UUID) error
}

type planService struct {
	transactor crud.Transactor
	repo       crud.Repository[entity.Plan]
}

func NewPlanService(
	transactor crud.Transactor,
	repo crud.Repository[entity.Plan],
) *planService {
	return &planService{
		transactor,
		repo,
	}
}

func (ps *planService) Create(ctx context.Context, req dto.NewPlanRequest) (dto.PlanResponse, error) {
	newPlan := entity.Plan{
		Name: req.Name,
	}

	insertedPlan, err := ps.repo.Insert(ctx, newPlan)
	if err != nil {
		return dto.PlanResponse{}, err
	}

	return mapper.PlanToResponse(insertedPlan), nil
}

func (ps *planService) GetList(ctx context.Context) ([]dto.PlanResponse, error) {
	spec := crud.Specification[entity.Plan]{}
	plans, err := ps.repo.FindAll(ctx, spec)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(plans, mapper.PlanToResponse), nil
}

func (ps *planService) GetOne(ctx context.Context, id uuid.UUID) (dto.PlanResponse, error) {
	spec := crud.Specification[entity.Plan]{}
	spec.Model.ID = id
	plan, err := ps.repo.FindFirst(ctx, spec)
	if err != nil {
		return dto.PlanResponse{}, err
	}
	if plan.IsZero() {
		return dto.PlanResponse{}, ungerr.NotFoundError("plan is not found")
	}

	return mapper.PlanToResponse(plan), nil
}

func (ps *planService) Update(ctx context.Context, req dto.UpdatePlanRequest) (dto.PlanResponse, error) {
	var resp dto.PlanResponse
	err := ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		spec := crud.Specification[entity.Plan]{}
		spec.Model.ID = req.ID
		spec.ForUpdate = true
		plan, err := ps.repo.FindFirst(ctx, spec)
		if err != nil {
			return err
		}
		if plan.IsZero() {
			return ungerr.NotFoundError("plan is not found")
		}

		plan.Name = req.Name
		plan.IsActive = req.IsActive

		updatedPlan, err := ps.repo.Update(ctx, plan)
		if err != nil {
			return err
		}

		resp = mapper.PlanToResponse(updatedPlan)
		return nil
	})
	return resp, err
}

func (ps *planService) Delete(ctx context.Context, id uuid.UUID) error {
	return ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		spec := crud.Specification[entity.Plan]{}
		spec.Model.ID = id
		spec.ForUpdate = true
		plan, err := ps.repo.FindFirst(ctx, spec)
		if err != nil {
			return err
		}
		if plan.IsZero() {
			return nil
		}

		return ps.Delete(ctx, id)
	})
}
