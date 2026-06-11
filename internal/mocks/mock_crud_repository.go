package mocks

import (
	"context"

	crud "github.com/itsLeonB/go-crud"
	mock "github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockRepository is a generic mock for crud.Repository[T].
type MockRepository[T any] struct {
	mock.Mock
}

func NewMockRepository[T any](t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRepository[T] {
	m := &MockRepository[T]{}
	m.Test(t)
	t.Cleanup(func() { m.AssertExpectations(t) })
	return m
}

func (m *MockRepository[T]) Insert(ctx context.Context, model T) (T, error) {
	ret := m.Called(ctx, model)
	return ret.Get(0).(T), ret.Error(1)
}

func (m *MockRepository[T]) FindAll(ctx context.Context, spec crud.Specification[T]) ([]T, error) {
	ret := m.Called(ctx, spec)
	return ret.Get(0).([]T), ret.Error(1)
}

func (m *MockRepository[T]) FindFirst(ctx context.Context, spec crud.Specification[T]) (T, error) {
	ret := m.Called(ctx, spec)
	return ret.Get(0).(T), ret.Error(1)
}

func (m *MockRepository[T]) Update(ctx context.Context, model T) (T, error) {
	ret := m.Called(ctx, model)
	return ret.Get(0).(T), ret.Error(1)
}

func (m *MockRepository[T]) Delete(ctx context.Context, model T) error {
	ret := m.Called(ctx, model)
	return ret.Error(0)
}

func (m *MockRepository[T]) InsertMany(ctx context.Context, models []T) ([]T, error) {
	ret := m.Called(ctx, models)
	return ret.Get(0).([]T), ret.Error(1)
}

func (m *MockRepository[T]) DeleteMany(ctx context.Context, models []T) error {
	ret := m.Called(ctx, models)
	return ret.Error(0)
}

func (m *MockRepository[T]) SaveMany(ctx context.Context, models []T) ([]T, error) {
	ret := m.Called(ctx, models)
	return ret.Get(0).([]T), ret.Error(1)
}

func (m *MockRepository[T]) GetGormInstance(ctx context.Context) (*gorm.DB, error) {
	ret := m.Called(ctx)
	return ret.Get(0).(*gorm.DB), ret.Error(1)
}
