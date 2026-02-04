package mocks

import (
	"context"
	"ticres/internal/entity"
	"github.com/stretchr/testify/mock"
)

type MockEventRepo struct{
	mock.Mock
}

func(m *MockEventRepo) CreateEvent(ctx context.Context, event *entity.Event) error{
	// m.called merekam bahwa funsi ini dipanggil dengan argument tertentu
	args := m.Called(ctx,event)

	// mengembalikan return value yang kita set sebelumnya (bisa nil atau error)
	return args.Error(0)
}
func(m *MockEventRepo) GetAllEvents(ctx context.Context) ([]entity.Event, error){
	args := m.Called(ctx)

	// Mengembalikan 2 value : []entity.Event dan error
	if args.Get(0) == nil {
		return nil ,args.Error(1)
	}

	return args.Get(0).([]entity.Event), args.Error(1)
}

func(m *MockEventRepo) GetEventByID(ctx context.Context, eventID int64) (*entity.Event, error) {
	args := m.Called(ctx)

	// Mengembalikan 2 value : []entity.Event dan error
	if args.Get(0) == nil {
		return nil ,args.Error(1)
	}

	return args.Get(0).(*entity.Event), args.Error(1)

}
func(m *MockEventRepo) UpdateEvent(ctx context.Context, event *entity.Event, preCapacity int64) error{
	args := m.Called(ctx,event)

	// mengembalikan return value yang kita set sebelumnya (bisa nil atau error)
	return args.Error(0)
}
func(m *MockEventRepo) UpdateEventStatus(ctx context.Context, eventID int64, status string) error{
	args := m.Called(ctx,eventID, status)

	// mengembalikan return value yang kita set sebelumnya (bisa nil atau error)
	return args.Error(0)
}