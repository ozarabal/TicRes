package mocks

import (
	"context"
	"ticres/internal/entity"
	"github.com/stretchr/testify/mock"
)


type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) CreateUser(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)

	return args.Error(0)
}

func (m *MockUserRepo) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepo) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	args := m.Called(ctx,id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entity.User), args.Error(1)
}