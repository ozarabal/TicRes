package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"ticres/internal/entity"
	"ticres/internal/usecase"
	"ticres/internal/usecase/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEventUsecase_CreateEvent(t *testing.T) {
	tests := []struct {
		name        string
		input       *entity.Event
		ticketPrice float64
		mock        func(mockRepo *mocks.MockEventRepo)
		wantErr     bool
	}{
		{
			name:        "Success Create Event",
			input:       &entity.Event{Name: "Konser Coldplay", Capacity: 1000},
			ticketPrice: 150000,
			mock: func(mockRepo *mocks.MockEventRepo) {
				mockRepo.On("CreateEvent", mock.Anything, mock.AnythingOfType("*entity.Event"), float64(150000)).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:        "Failed Create Event - DB Error",
			input:       &entity.Event{Name: "Konser B", Capacity: 100},
			ticketPrice: 50000,
			mock: func(mockRepo *mocks.MockEventRepo) {
				mockRepo.On("CreateEvent", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockEventRepo)
			mockNotif := new(mocks.MockNotificationService)

			tt.mock(mockRepo)

			u := usecase.NewEventUsecase(mockRepo, time.Second*2, mockNotif)
			err := u.CreateEvent(context.Background(), tt.input, tt.ticketPrice)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestEventUsecase_ListEvents(t *testing.T) {
	mockEvents := []entity.Event{
		{ID: 1, Name: "Konser A", Location: "Jakarta", Capacity: 1000},
		{ID: 2, Name: "Konser B", Location: "Bandung", Capacity: 500},
	}

	tests := []struct {
		name       string
		mock       func(mockRepo *mocks.MockEventRepo)
		wantErr    bool
		wantEvents []entity.Event
	}{
		{
			name: "Success List Events",
			mock: func(mockRepo *mocks.MockEventRepo) {
				mockRepo.On("GetAllEvents", mock.Anything).Return(mockEvents, nil).Once()
			},
			wantErr:    false,
			wantEvents: mockEvents,
		},
		{
			name: "Success List Events - Empty",
			mock: func(mockRepo *mocks.MockEventRepo) {
				mockRepo.On("GetAllEvents", mock.Anything).Return([]entity.Event{}, nil).Once()
			},
			wantErr:    false,
			wantEvents: []entity.Event{},
		},
		{
			name: "Failed List Events - DB Error",
			mock: func(mockRepo *mocks.MockEventRepo) {
				mockRepo.On("GetAllEvents", mock.Anything).Return(nil, errors.New("db error")).Once()
			},
			wantErr:    true,
			wantEvents: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockEventRepo)
			mockNotif := new(mocks.MockNotificationService)

			tt.mock(mockRepo)

			u := usecase.NewEventUsecase(mockRepo, time.Second*2, mockNotif)
			events, err := u.ListEvents(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, events)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantEvents, events)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestEventUsecase_ListEventsWithSearch(t *testing.T) {
	mockEvents := []entity.Event{
		{ID: 1, Name: "Konser Coldplay", Location: "Jakarta", Capacity: 1000},
		{ID: 2, Name: "Konser Westlife", Location: "Bandung", Capacity: 500},
	}

	tests := []struct {
		name       string
		search     string
		page       int
		limit      int
		mock       func(mockRepo *mocks.MockEventRepo)
		wantErr    bool
		wantEvents []entity.Event
		wantTotal  int
	}{
		{
			name:   "Success - Search with Results",
			search: "Konser",
			page:   1,
			limit:  10,
			mock: func(mockRepo *mocks.MockEventRepo) {
				mockRepo.On("GetEventsWithSearch", mock.Anything, "Konser", 1, 10).
					Return(mockEvents, 2, nil).Once()
			},
			wantErr:    false,
			wantEvents: mockEvents,
			wantTotal:  2,
		},
		{
			name:   "Success - Search Empty Result",
			search: "NonExistent",
			page:   1,
			limit:  10,
			mock: func(mockRepo *mocks.MockEventRepo) {
				mockRepo.On("GetEventsWithSearch", mock.Anything, "NonExistent", 1, 10).
					Return([]entity.Event{}, 0, nil).Once()
			},
			wantErr:    false,
			wantEvents: []entity.Event{},
			wantTotal:  0,
		},
		{
			name:   "Success - Pagination Page 2",
			search: "",
			page:   2,
			limit:  1,
			mock: func(mockRepo *mocks.MockEventRepo) {
				mockRepo.On("GetEventsWithSearch", mock.Anything, "", 2, 1).
					Return(mockEvents[1:], 2, nil).Once()
			},
			wantErr:    false,
			wantEvents: mockEvents[1:],
			wantTotal:  2,
		},
		{
			name:   "Failed - DB Error",
			search: "Konser",
			page:   1,
			limit:  10,
			mock: func(mockRepo *mocks.MockEventRepo) {
				mockRepo.On("GetEventsWithSearch", mock.Anything, "Konser", 1, 10).
					Return(nil, 0, errors.New("db error")).Once()
			},
			wantErr:    true,
			wantEvents: nil,
			wantTotal:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockEventRepo)
			mockNotif := new(mocks.MockNotificationService)

			tt.mock(mockRepo)

			u := usecase.NewEventUsecase(mockRepo, time.Second*2, mockNotif)
			events, total, err := u.ListEventsWithSearch(context.Background(), tt.search, tt.page, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, events)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantEvents, events)
				assert.Equal(t, tt.wantTotal, total)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestEventUsecase_GetEventByID(t *testing.T) {
	mockEvent := &entity.Event{ID: 1, Name: "Konser Coldplay", Location: "Jakarta", Capacity: 1000}

	tests := []struct {
		name      string
		eventID   int64
		mock      func(mockRepo *mocks.MockEventRepo)
		wantErr   bool
		wantEvent *entity.Event
	}{
		{
			name:    "Success - Get Event",
			eventID: 1,
			mock: func(mockRepo *mocks.MockEventRepo) {
				mockRepo.On("GetEventByID", mock.Anything, int64(1)).
					Return(mockEvent, nil).Once()
			},
			wantErr:   false,
			wantEvent: mockEvent,
		},
		{
			name:    "Failed - Not Found",
			eventID: 999,
			mock: func(mockRepo *mocks.MockEventRepo) {
				mockRepo.On("GetEventByID", mock.Anything, int64(999)).
					Return(nil, entity.ErrNotFound).Once()
			},
			wantErr:   true,
			wantEvent: nil,
		},
		{
			name:    "Failed - DB Error",
			eventID: 1,
			mock: func(mockRepo *mocks.MockEventRepo) {
				mockRepo.On("GetEventByID", mock.Anything, int64(1)).
					Return(nil, errors.New("db error")).Once()
			},
			wantErr:   true,
			wantEvent: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockEventRepo)
			mockNotif := new(mocks.MockNotificationService)

			tt.mock(mockRepo)

			u := usecase.NewEventUsecase(mockRepo, time.Second*2, mockNotif)
			event, err := u.GetEventByID(context.Background(), tt.eventID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, event)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantEvent, event)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestEventUsecase_GetEventWithSeats(t *testing.T) {
	mockEventWithSeats := &entity.EventWithSeats{
		Event: entity.Event{ID: 1, Name: "Konser Coldplay", Location: "Jakarta", Capacity: 3},
		Seats: []entity.Seat{
			{ID: 1, EventID: 1, SeatNumber: "1-1", IsBooked: false},
			{ID: 2, EventID: 1, SeatNumber: "1-2", IsBooked: true},
			{ID: 3, EventID: 1, SeatNumber: "1-3", IsBooked: false},
		},
	}

	tests := []struct {
		name      string
		eventID   int64
		mock      func(mockRepo *mocks.MockEventRepo)
		wantErr   bool
		wantEvent *entity.EventWithSeats
	}{
		{
			name:    "Success - Get Event With Seats",
			eventID: 1,
			mock: func(mockRepo *mocks.MockEventRepo) {
				mockRepo.On("GetEventWithSeats", mock.Anything, int64(1)).
					Return(mockEventWithSeats, nil).Once()
			},
			wantErr:   false,
			wantEvent: mockEventWithSeats,
		},
		{
			name:    "Failed - Not Found",
			eventID: 999,
			mock: func(mockRepo *mocks.MockEventRepo) {
				mockRepo.On("GetEventWithSeats", mock.Anything, int64(999)).
					Return(nil, entity.ErrNotFound).Once()
			},
			wantErr:   true,
			wantEvent: nil,
		},
		{
			name:    "Failed - DB Error",
			eventID: 1,
			mock: func(mockRepo *mocks.MockEventRepo) {
				mockRepo.On("GetEventWithSeats", mock.Anything, int64(1)).
					Return(nil, errors.New("db error")).Once()
			},
			wantErr:   true,
			wantEvent: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockEventRepo)
			mockNotif := new(mocks.MockNotificationService)

			tt.mock(mockRepo)

			u := usecase.NewEventUsecase(mockRepo, time.Second*2, mockNotif)
			eventWithSeats, err := u.GetEventWithSeats(context.Background(), tt.eventID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, eventWithSeats)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantEvent, eventWithSeats)
				assert.Equal(t, 3, len(eventWithSeats.Seats))
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestEventUsecase_EditEvent(t *testing.T) {
	tests := []struct {
		name        string
		input       *entity.Event
		prevCapacity int64
		mock        func(mockRepo *mocks.MockEventRepo)
		wantErr     bool
	}{
		{
			name:        "Success Edit Event",
			input:       &entity.Event{ID: 1, Name: "Konser Updated", Capacity: 2000},
			prevCapacity: 1000,
			mock: func(mockRepo *mocks.MockEventRepo) {
				mockRepo.On("UpdateEvent", mock.Anything, mock.AnythingOfType("*entity.Event")).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:        "Failed Edit Event - Not Found",
			input:       &entity.Event{ID: 999, Name: "Konser Unknown", Capacity: 100},
			prevCapacity: 100,
			mock: func(mockRepo *mocks.MockEventRepo) {
				mockRepo.On("UpdateEvent", mock.Anything, mock.Anything).Return(entity.ErrNotFound).Once()
			},
			wantErr: true,
		},
		{
			name:        "Failed Edit Event - DB Error",
			input:       &entity.Event{ID: 1, Name: "Konser Error", Capacity: 500},
			prevCapacity: 1000,
			mock: func(mockRepo *mocks.MockEventRepo) {
				mockRepo.On("UpdateEvent", mock.Anything, mock.Anything).Return(errors.New("db error")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockEventRepo)
			mockNotif := new(mocks.MockNotificationService)

			tt.mock(mockRepo)

			u := usecase.NewEventUsecase(mockRepo, time.Second*2, mockNotif)
			err := u.EditEvent(context.Background(), tt.input, tt.prevCapacity)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestEventUsecase_CancelEvent(t *testing.T) {
	tests := []struct {
		name    string
		eventID int64
		mock    func(mockRepo *mocks.MockEventRepo, mockNotif *mocks.MockNotificationService)
		wantErr bool
	}{
		{
			name:    "Success Cancel Event",
			eventID: 1,
			mock: func(mockRepo *mocks.MockEventRepo, mockNotif *mocks.MockNotificationService) {
				mockRepo.On("UpdateEventStatus", mock.Anything, int64(1), "CANCELLED").Return(nil).Once()
				mockNotif.On("EnqueueCancellation", int64(1)).Once()
			},
			wantErr: false,
		},
		{
			name:    "Failed Cancel Event - Not Found",
			eventID: 999,
			mock: func(mockRepo *mocks.MockEventRepo, mockNotif *mocks.MockNotificationService) {
				mockRepo.On("UpdateEventStatus", mock.Anything, int64(999), "CANCELLED").Return(entity.ErrNotFound).Once()
			},
			wantErr: true,
		},
		{
			name:    "Failed Cancel Event - DB Error",
			eventID: 1,
			mock: func(mockRepo *mocks.MockEventRepo, mockNotif *mocks.MockNotificationService) {
				mockRepo.On("UpdateEventStatus", mock.Anything, int64(1), "CANCELLED").Return(errors.New("db error")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockEventRepo)
			mockNotif := new(mocks.MockNotificationService)

			tt.mock(mockRepo, mockNotif)

			u := usecase.NewEventUsecase(mockRepo, time.Second*2, mockNotif)
			err := u.CancelEvent(context.Background(), tt.eventID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
			mockNotif.AssertExpectations(t)
		})
	}
}
