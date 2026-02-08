package mocks

import (
	"context"

	"ticres/internal/entity"

	"github.com/stretchr/testify/mock"
)

type MockEventRepo struct {
	mock.Mock
}

func (m *MockEventRepo) CreateEvent(ctx context.Context, event *entity.Event, ticketPrice float64) error {
	args := m.Called(ctx, event, ticketPrice)
	return args.Error(0)
}

func (m *MockEventRepo) GetAllEvents(ctx context.Context) ([]entity.Event, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Event), args.Error(1)
}

func (m *MockEventRepo) GetEventsWithSearch(ctx context.Context, search string, page, limit int) ([]entity.Event, int, error) {
	args := m.Called(ctx, search, page, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]entity.Event), args.Int(1), args.Error(2)
}

func (m *MockEventRepo) GetEventByID(ctx context.Context, eventID int64) (*entity.Event, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Event), args.Error(1)
}

func (m *MockEventRepo) GetEventWithSeats(ctx context.Context, eventID int64) (*entity.EventWithSeats, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.EventWithSeats), args.Error(1)
}

func (m *MockEventRepo) GetSeatsByEventID(ctx context.Context, eventID int64) ([]entity.Seat, error) {
	args := m.Called(ctx, eventID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Seat), args.Error(1)
}

func (m *MockEventRepo) UpdateEvent(ctx context.Context, event *entity.Event, preCapacity int64) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventRepo) UpdateEventStatus(ctx context.Context, eventID int64, status string) error {
	args := m.Called(ctx, eventID, status)
	return args.Error(0)
}
