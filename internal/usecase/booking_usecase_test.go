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

func TestBookingUsecase_BookSeats(t *testing.T) {
	tests := []struct {
		name      string
		userID    int64
		eventID   int64
		seatIDs   []int64
		userEmail string
		mock      func(mockRepo *mocks.MockBookingRepo, mockTxnRepo *mocks.MockTransactionRepo, mockNotif *mocks.MockNotificationService)
		wantErr   bool
	}{
		{
			name:      "Success Booking",
			userID:    1,
			eventID:   10,
			seatIDs:   []int64{101, 102},
			userEmail: "user@test.com",
			mock: func(mockRepo *mocks.MockBookingRepo, mockTxnRepo *mocks.MockTransactionRepo, mockNotif *mocks.MockNotificationService) {
				mockRepo.On("CreateBooking", mock.Anything, int64(1), int64(10), []int64{101, 102}).
					Return(int64(999), float64(200000), nil).Once()
				mockTxnRepo.On("CreateTransaction", mock.Anything, mock.AnythingOfType("*entity.Transaction")).
					Return(nil).Once()
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
			mock: func(mockRepo *mocks.MockBookingRepo, mockTxnRepo *mocks.MockTransactionRepo, mockNotif *mocks.MockNotificationService) {
				mockRepo.On("CreateBooking", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(int64(0), float64(0), errors.New("seat not available")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockBookingRepo)
			mockTxnRepo := new(mocks.MockTransactionRepo)
			mockNotif := new(mocks.MockNotificationService)

			tt.mock(mockRepo, mockTxnRepo, mockNotif)

			u := usecase.NewBookingUsecase(mockRepo, mockTxnRepo, time.Second*2, mockNotif)
			result, err := u.BookSeats(context.Background(), tt.userID, tt.eventID, tt.seatIDs, tt.userEmail)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "PENDING", result.Status)
				assert.Equal(t, float64(200000), result.TotalAmount)
			}

			mockRepo.AssertExpectations(t)
			mockTxnRepo.AssertExpectations(t)
			mockNotif.AssertExpectations(t)
		})
	}
}

func TestBookingUsecase_GetBookingsByUserID(t *testing.T) {
	now := time.Now()
	mockBookings := []entity.BookingWithDetails{
		{ID: 1, UserID: 1, UserName: "John", UserEmail: "john@test.com", EventID: 10, EventName: "Concert A", Status: "PAID", CreatedAt: now},
		{ID: 2, UserID: 1, UserName: "John", UserEmail: "john@test.com", EventID: 20, EventName: "Concert B", Status: "PENDING", CreatedAt: now},
	}

	tests := []struct {
		name         string
		userID       int64
		mock         func(mockRepo *mocks.MockBookingRepo)
		wantErr      bool
		wantBookings []entity.BookingWithDetails
	}{
		{
			name:   "Success - Get User Bookings",
			userID: 1,
			mock: func(mockRepo *mocks.MockBookingRepo) {
				mockRepo.On("GetBookingsByUserID", mock.Anything, int64(1)).
					Return(mockBookings, nil).Once()
			},
			wantErr:      false,
			wantBookings: mockBookings,
		},
		{
			name:   "Success - No Bookings",
			userID: 2,
			mock: func(mockRepo *mocks.MockBookingRepo) {
				mockRepo.On("GetBookingsByUserID", mock.Anything, int64(2)).
					Return([]entity.BookingWithDetails{}, nil).Once()
			},
			wantErr:      false,
			wantBookings: []entity.BookingWithDetails{},
		},
		{
			name:   "Failed - DB Error",
			userID: 1,
			mock: func(mockRepo *mocks.MockBookingRepo) {
				mockRepo.On("GetBookingsByUserID", mock.Anything, int64(1)).
					Return(nil, errors.New("db error")).Once()
			},
			wantErr:      true,
			wantBookings: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockBookingRepo)
			mockTxnRepo := new(mocks.MockTransactionRepo)
			mockNotif := new(mocks.MockNotificationService)

			tt.mock(mockRepo)

			u := usecase.NewBookingUsecase(mockRepo, mockTxnRepo, time.Second*2, mockNotif)
			bookings, err := u.GetBookingsByUserID(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, bookings)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantBookings, bookings)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBookingUsecase_GetAllBookings(t *testing.T) {
	mockBookings := []entity.BookingWithDetails{
		{ID: 1, UserID: 1, UserName: "John", UserEmail: "john@test.com", EventID: 10, EventName: "Concert A", Status: "PAID"},
		{ID: 2, UserID: 2, UserName: "Jane", UserEmail: "jane@test.com", EventID: 10, EventName: "Concert A", Status: "PENDING"},
	}

	tests := []struct {
		name         string
		status       string
		sortBy       string
		sortOrder    string
		page         int
		limit        int
		mock         func(mockRepo *mocks.MockBookingRepo)
		wantErr      bool
		wantBookings []entity.BookingWithDetails
		wantTotal    int
	}{
		{
			name:      "Success - Get All Bookings",
			status:    "",
			sortBy:    "created_at",
			sortOrder: "desc",
			page:      1,
			limit:     20,
			mock: func(mockRepo *mocks.MockBookingRepo) {
				mockRepo.On("GetAllBookings", mock.Anything, "", "created_at", "desc", 1, 20).
					Return(mockBookings, 2, nil).Once()
			},
			wantErr:      false,
			wantBookings: mockBookings,
			wantTotal:    2,
		},
		{
			name:      "Failed - DB Error",
			status:    "",
			sortBy:    "created_at",
			sortOrder: "desc",
			page:      1,
			limit:     20,
			mock: func(mockRepo *mocks.MockBookingRepo) {
				mockRepo.On("GetAllBookings", mock.Anything, "", "created_at", "desc", 1, 20).
					Return(nil, 0, errors.New("db error")).Once()
			},
			wantErr:      true,
			wantBookings: nil,
			wantTotal:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockBookingRepo)
			mockTxnRepo := new(mocks.MockTransactionRepo)
			mockNotif := new(mocks.MockNotificationService)

			tt.mock(mockRepo)

			u := usecase.NewBookingUsecase(mockRepo, mockTxnRepo, time.Second*2, mockNotif)
			bookings, total, err := u.GetAllBookings(context.Background(), tt.status, tt.sortBy, tt.sortOrder, tt.page, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, bookings)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantBookings, bookings)
				assert.Equal(t, tt.wantTotal, total)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBookingUsecase_GetBookingsByEventID(t *testing.T) {
	mockBookings := []entity.BookingWithDetails{
		{ID: 1, UserID: 1, UserName: "John", UserEmail: "john@test.com", EventID: 10, EventName: "Concert A", Status: "PAID"},
		{ID: 2, UserID: 2, UserName: "Jane", UserEmail: "jane@test.com", EventID: 10, EventName: "Concert A", Status: "PENDING"},
	}

	tests := []struct {
		name         string
		eventID      int64
		status       string
		sortBy       string
		sortOrder    string
		mock         func(mockRepo *mocks.MockBookingRepo)
		wantErr      bool
		wantBookings []entity.BookingWithDetails
	}{
		{
			name:      "Success - Get Event Bookings",
			eventID:   10,
			status:    "",
			sortBy:    "created_at",
			sortOrder: "desc",
			mock: func(mockRepo *mocks.MockBookingRepo) {
				mockRepo.On("GetBookingsWithDetailsByEventID", mock.Anything, int64(10), "", "created_at", "desc").
					Return(mockBookings, nil).Once()
			},
			wantErr:      false,
			wantBookings: mockBookings,
		},
		{
			name:      "Failed - DB Error",
			eventID:   10,
			status:    "",
			sortBy:    "created_at",
			sortOrder: "desc",
			mock: func(mockRepo *mocks.MockBookingRepo) {
				mockRepo.On("GetBookingsWithDetailsByEventID", mock.Anything, int64(10), "", "created_at", "desc").
					Return(nil, errors.New("db error")).Once()
			},
			wantErr:      true,
			wantBookings: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockBookingRepo)
			mockTxnRepo := new(mocks.MockTransactionRepo)
			mockNotif := new(mocks.MockNotificationService)

			tt.mock(mockRepo)

			u := usecase.NewBookingUsecase(mockRepo, mockTxnRepo, time.Second*2, mockNotif)
			bookings, err := u.GetBookingsByEventID(context.Background(), tt.eventID, tt.status, tt.sortBy, tt.sortOrder)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, bookings)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantBookings, bookings)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
