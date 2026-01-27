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


func TestEventUsecase_CreteEvent(t *testing.T){

	mockRepo := new(mocks.MockEventRepo)

	u := usecase.NewEventUsecase(mockRepo, time.Second*2)

	tests := []struct {
		name string
		input *entity.Event
		mock	func()
		wantErr bool
	}{
		{
			name: "Success Create Event",
			input: &entity.Event{Name: "Konser coldp", Capacity: 1000},
			mock: func ()  {
				mockRepo.On("CreateEvent", mock.Anything, mock.AnythingOfType("*entity.Event")).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "Failed Create Event - DB Error",
			input: &entity.Event{Name: "Konser B", Capacity: 100},
			mock: func() {
				mockRepo.On("CreateEvent", mock.Anything, mock.Anything).Return(errors.New("db error")).Once()
			},
			wantErr: true,
		},
	}

	for _,tt := range tests {
		t.Run(tt.name, func(t *testing.T){
			tt.mock()
			err := u.CreateEvent(context.Background(), tt.input)
			if tt.wantErr{
				assert.Error(t, err)
			} else{
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
	
}