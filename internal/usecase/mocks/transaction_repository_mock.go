package mocks

import (
	"context"
	"ticres/internal/entity"

	"github.com/stretchr/testify/mock"
)

type MockTransactionRepo struct {
	mock.Mock
}

func (m *MockTransactionRepo) CreateTransaction(ctx context.Context, txn *entity.Transaction) error {
	args := m.Called(ctx, txn)
	return args.Error(0)
}

func (m *MockTransactionRepo) GetTransactionByBookingID(ctx context.Context, bookingID int64) (*entity.Transaction, error) {
	args := m.Called(ctx, bookingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Transaction), args.Error(1)
}

func (m *MockTransactionRepo) GetTransactionByExternalID(ctx context.Context, externalID string) (*entity.Transaction, error) {
	args := m.Called(ctx, externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Transaction), args.Error(1)
}

func (m *MockTransactionRepo) UpdateTransactionStatus(ctx context.Context, paymentID int64, status, externalID string) error {
	args := m.Called(ctx, paymentID, status, externalID)
	return args.Error(0)
}
