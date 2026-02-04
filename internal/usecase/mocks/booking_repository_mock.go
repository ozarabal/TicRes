package mocks

import (
	"context"
	"ticres/internal/entity"

	"github.com/stretchr/testify/mock"
)

type MockBookingRepo struct {
	mock.Mock
}

func (m *MockBookingRepo) CreateBooking(ctx context.Context, userID, eventID int64, seatIDs []int64) (int64, error){
	args := m.Called(ctx, userID, eventID, seatIDs)

	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBookingRepo) GetBookingsByEventID(ctx context.Context, eventID int64) ([]entity.Booking, error){
	args := m.Called(ctx, eventID)

	return args.Get(0).([]entity.Booking), args.Error(1)
}

func (m *MockBookingRepo) UpdateBookingStatus(ctx context.Context, bookingID int64, status string) error{
	args := m.Called(ctx, bookingID, status)

	return args.Error(0)
}