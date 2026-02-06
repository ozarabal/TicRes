package usecase

import (
	"context"
	"time"

	"ticres/internal/entity"
	"ticres/internal/repository"
)

type BookingUsecase interface {
	BookSeats(ctx context.Context, userID, eventID int64, seatIDs []int64, userEmail string) error
	GetBookingsByUserID(ctx context.Context, userID int64) ([]entity.BookingWithDetails, error)
	GetAllBookings(ctx context.Context, status, sortBy, sortOrder string, page, limit int) ([]entity.BookingWithDetails, int, error)
	GetBookingsByEventID(ctx context.Context, eventID int64, status, sortBy, sortOrder string) ([]entity.BookingWithDetails, error)
}

type NotificationService interface {
	SendNotification(bookingID int64, email, message string)
	EnqueueCancellation(eventID int64)
}

type bookingUsecase struct {
	bookingRepo    repository.BookingRepository
	contextTimeout time.Duration
	// notifWorker    *worker.NotificationWorker
	notifWorker    NotificationService

}

func NewBookingUsecase(repo repository.BookingRepository, timeout time.Duration, notifWorker NotificationService) BookingUsecase {
	return &bookingUsecase{bookingRepo: repo, contextTimeout: timeout, notifWorker: notifWorker}
}

func (uc *bookingUsecase) BookSeats(ctx context.Context, userID, eventID int64, seatIDs []int64, useremail string) error {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	bookingID, err := uc.bookingRepo.CreateBooking(ctx, userID, eventID, seatIDs)
	if err != nil {
		return err
	}

	uc.notifWorker.SendNotification(bookingID, useremail, "Booking Berhasil!")

	return nil
}

func (uc *bookingUsecase) GetBookingsByUserID(ctx context.Context, userID int64) ([]entity.BookingWithDetails, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	return uc.bookingRepo.GetBookingsByUserID(ctx, userID)
}

func (uc *bookingUsecase) GetAllBookings(ctx context.Context, status, sortBy, sortOrder string, page, limit int) ([]entity.BookingWithDetails, int, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	return uc.bookingRepo.GetAllBookings(ctx, status, sortBy, sortOrder, page, limit)
}

func (uc *bookingUsecase) GetBookingsByEventID(ctx context.Context, eventID int64, status, sortBy, sortOrder string) ([]entity.BookingWithDetails, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	return uc.bookingRepo.GetBookingsWithDetailsByEventID(ctx, eventID, status, sortBy, sortOrder)
}