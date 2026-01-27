package mocks

import(

	"context"
	"github.com/stretchr/testify/mock"
)

type MockBookingRepo struct {
	mock.Mock
}

func (m *MockBookingRepo) CreateBooking(ctx context.Context, userID, eventID int64, seatIDs []int64) (int64, error){
	args := m.Called(ctx, userID, eventID, seatIDs)

	return args.Get(0).(int64), args.Error(1)
}