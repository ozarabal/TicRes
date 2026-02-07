package mocks

import (
	"context"
	"ticres/internal/entity"

	"github.com/stretchr/testify/mock"
)

type MockBookingRepo struct {
	mock.Mock
}

func (m *MockBookingRepo) CreateBooking(ctx context.Context, userID, eventID int64, seatIDs []int64) (int64, error) {
	args := m.Called(ctx, userID, eventID, seatIDs)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBookingRepo) GetBookingsByEventID(ctx context.Context, eventID int64) ([]entity.Booking, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Booking), args.Error(1)
}

func (m *MockBookingRepo) GetBookingsByUserID(ctx context.Context, userID int64) ([]entity.BookingWithDetails, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.BookingWithDetails), args.Error(1)
}

func (m *MockBookingRepo) GetAllBookings(ctx context.Context, status, sortBy, sortOrder string, page, limit int) ([]entity.BookingWithDetails, int, error) {
	args := m.Called(ctx, status, sortBy, sortOrder, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.BookingWithDetails), args.Int(1), args.Error(2)
}

func (m *MockBookingRepo) GetBookingsWithDetailsByEventID(ctx context.Context, eventID int64, status, sortBy, sortOrder string) ([]entity.BookingWithDetails, error) {
	args := m.Called(ctx, eventID, status, sortBy, sortOrder)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.BookingWithDetails), args.Error(1)
}

func (m *MockBookingRepo) UpdateBookingStatus(ctx context.Context, bookingID int64, status string) error {
	args := m.Called(ctx, bookingID, status)
	return args.Error(0)
}
