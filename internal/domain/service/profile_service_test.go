package service_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/cashback/internal/mocks"
	"github.com/itsLeonB/go-crud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestProfileService(
	t *testing.T,
) (service.ProfileService, *mocks.MockProfileRepository, *mocks.MockRepository[users.User], *mocks.MockRepository[users.RelatedProfile]) {
	profileRepo := mocks.NewMockProfileRepository(t)
	userRepo := mocks.NewMockRepository[users.User](t)
	transactor := mocks.NewMockTransactor(t)
	friendshipRepo := mocks.NewMockFriendshipRepository(t)
	relatedProfileRepo := mocks.NewMockRepository[users.RelatedProfile](t)
	subLimitSvc := mocks.NewMockSubscriptionLimitService(t)

	svc := service.NewProfileService(
		transactor,
		profileRepo,
		userRepo,
		friendshipRepo,
		relatedProfileRepo,
		nil,
		subLimitSvc,
	)

	return svc, profileRepo, userRepo, relatedProfileRepo
}

func TestSearch_ByName_ReturnsOnlyMinimalFields(t *testing.T) {
	svc, profileRepo, _, _ := newTestProfileService(t)

	callerID := uuid.New()
	friendID := uuid.New()

	profileRepo.EXPECT().SearchByName(mock.Anything, "alice", 10).Return([]users.UserProfile{
		{
			BaseEntity:   crud.BaseEntity{ID: friendID},
			UserID:       uuid.NullUUID{UUID: uuid.New(), Valid: true},
			Name:         "Alice",
			Avatar:       "https://img.example.com/alice.png",
			HomeCurrency: "USD",
		},
	}, nil)

	results, err := svc.Search(context.Background(), callerID, "alice")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, dto.SearchProfileResponse{
		ID:     friendID,
		Name:   "Alice",
		Avatar: "https://img.example.com/alice.png",
	}, results[0])
}

func TestSearch_ByName_ExcludesSelf(t *testing.T) {
	svc, profileRepo, _, _ := newTestProfileService(t)

	callerID := uuid.New()

	profileRepo.EXPECT().SearchByName(mock.Anything, "test", 10).Return([]users.UserProfile{
		{
			BaseEntity: crud.BaseEntity{ID: callerID},
			Name:       "Test User",
			Avatar:     "avatar.png",
		},
	}, nil)

	results, err := svc.Search(context.Background(), callerID, "test")

	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestSearch_ByEmail_ReturnsOnlyMinimalFields(t *testing.T) {
	svc, _, userRepo, relatedProfileRepo := newTestProfileService(t)

	callerID := uuid.New()
	friendID := uuid.New()
	friendUserID := uuid.New()

	userRepo.On("FindFirst", mock.Anything, mock.Anything).Return(users.User{
		BaseEntity: crud.BaseEntity{ID: friendUserID},
		Email:      "alice@example.com",
		VerifiedAt: sql.NullTime{Time: time.Now(), Valid: true},
		Profile: users.UserProfile{
			BaseEntity:   crud.BaseEntity{ID: friendID},
			UserID:       uuid.NullUUID{UUID: friendUserID, Valid: true},
			Name:         "Alice",
			Avatar:       "https://img.example.com/alice.png",
			HomeCurrency: "USD",
		},
	}, nil)

	relatedProfileRepo.On("FindAll", mock.Anything, mock.Anything).Return([]users.RelatedProfile{}, nil)

	results, err := svc.Search(context.Background(), callerID, "alice@example.com")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, dto.SearchProfileResponse{
		ID:     friendID,
		Name:   "Alice",
		Avatar: "https://img.example.com/alice.png",
	}, results[0])
}

func TestSearch_ByEmail_ExcludesSelf(t *testing.T) {
	svc, _, userRepo, relatedProfileRepo := newTestProfileService(t)

	callerID := uuid.New()
	userID := uuid.New()

	userRepo.On("FindFirst", mock.Anything, mock.Anything).Return(users.User{
		BaseEntity: crud.BaseEntity{ID: userID},
		Email:      "me@example.com",
		VerifiedAt: sql.NullTime{Time: time.Now(), Valid: true},
		Profile: users.UserProfile{
			BaseEntity: crud.BaseEntity{ID: callerID},
			UserID:     uuid.NullUUID{UUID: userID, Valid: true},
			Name:       "Me",
			Avatar:     "me.png",
		},
	}, nil)

	relatedProfileRepo.On("FindAll", mock.Anything, mock.Anything).Return([]users.RelatedProfile{}, nil)

	results, err := svc.Search(context.Background(), callerID, "me@example.com")

	assert.NoError(t, err)
	assert.Empty(t, results)
}
