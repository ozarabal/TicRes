package mocks

import (
	"context"
	"ticres/internal/entity"

	"github.com/stretchr/testify/mock"
)

type MockRefundRepo struct {
	mock.Mock
}

func (m *MockRefundRepo) CreateRefund(ctx context.Context, refund *entity.Refund) error {
	args := m.Called(ctx, refund)
	return args.Error(0)
}

func (m *MockRefundRepo) GetRefundByBookingID(ctx context.Context, bookingID int64) (*entity.Refund, error) {
	args := m.Called(ctx, bookingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Refund), args.Error(1)
}
