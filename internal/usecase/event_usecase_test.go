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
		name    string
		input   *entity.Event
		mock    func(mockRepo *mocks.MockEventRepo)
		wantErr bool
	}{
		{
			name:  "Success Create Event",
			input: &entity.Event{Name: "Konser Coldplay", Capacity: 1000},
			mock: func(mockRepo *mocks.MockEventRepo) {
				mockRepo.On("CreateEvent", mock.Anything, mock.AnythingOfType("*entity.Event")).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:  "Failed Create Event - DB Error",
			input: &entity.Event{Name: "Konser B", Capacity: 100},
			mock: func(mockRepo *mocks.MockEventRepo) {
				mockRepo.On("CreateEvent", mock.Anything, mock.Anything).Return(errors.New("db error")).Once()
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
			err := u.CreateEvent(context.Background(), tt.input)

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
