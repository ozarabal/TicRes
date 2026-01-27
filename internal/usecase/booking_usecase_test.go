package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"ticres/internal/usecase"
	"ticres/internal/usecase/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBookingUsecase_BookSeats(t *testing.T) {
	// Setup Mocks
	mockRepo := new(mocks.MockBookingRepo)
	mockNotif := new(mocks.MockNotificationService) // Mock Worker

	// Init Usecase dengan Mock Worker
	u := usecase.NewBookingUsecase(mockRepo, time.Second*2, mockNotif)

	tests := []struct {
		name      string
		userID    int64
		eventID   int64
		seatIDs   []int64
		userEmail string
		mock      func()
		wantErr   bool
	}{
		{
			name:      "Success Booking",
			userID:    1,
			eventID:   10,
			seatIDs:   []int64{101, 102},
			userEmail: "user@test.com",
			mock: func() {
				// 1. Repo CreateBooking Berhasil -> Return BookingID 999
				mockRepo.On("CreateBooking", mock.Anything, int64(1), int64(10), []int64{101, 102}).
					Return(int64(999), nil).Once()

				// 2. KARENA SUKSES, Worker SendNotification HARUS dipanggil
				mockNotif.On("SendNotification", int64(999), "user@test.com", mock.AnythingOfType("string")).
					Once()
			},
			wantErr: false,
		},
		{
			name:      "Failed Booking - Seats Taken",
			userID:    1,
			eventID:   10,
			seatIDs:   []int64{101},
			userEmail: "user@test.com",
			mock: func() {
				// 1. Repo Gagal (Kursi penuh)
				mockRepo.On("CreateBooking", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(int64(0), errors.New("seat not available")).Once()

				// 2. Worker JANGAN dipanggil (Karena booking gagal)
				// Tidak perlu define mockNotif.On(...) karena ekspektasinya tidak dipanggil.
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset Mocks (Penting jika diluar loop)
			mockRepo.ExpectedCalls = nil
			mockNotif.ExpectedCalls = nil
			
			tt.mock()

			err := u.BookSeats(context.Background(), tt.userID, tt.eventID, tt.seatIDs, tt.userEmail)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verifikasi bahwa Mock terpanggil sesuai skenario
			mockRepo.AssertExpectations(t)
			mockNotif.AssertExpectations(t)
		})
	}
}