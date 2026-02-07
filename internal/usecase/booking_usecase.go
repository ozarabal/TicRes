package usecase

import (
	"context"
	"time"

	"ticres/internal/entity"
	"ticres/internal/repository"
	"ticres/pkg/logger"
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
	logger.Debug("usecase: booking seats",
		logger.Int64("user_id", userID),
		logger.Int64("event_id", eventID),
		logger.Int("seat_count", len(seatIDs)),
	)

	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	bookingID, err := uc.bookingRepo.CreateBooking(ctx, userID, eventID, seatIDs)
	if err != nil {
		logger.Error("usecase: failed to book seats",
			logger.Int64("user_id", userID),
			logger.Int64("event_id", eventID),
			logger.Err(err),
		)
		return err
	}

	uc.notifWorker.SendNotification(bookingID, useremail, "Booking Berhasil!")

	logger.Info("usecase: seats booked successfully",
		logger.Int64("booking_id", bookingID),
		logger.Int64("user_id", userID),
		logger.Int64("event_id", eventID),
	)
	return nil
}

func (uc *bookingUsecase) GetBookingsByUserID(ctx context.Context, userID int64) ([]entity.BookingWithDetails, error) {
	logger.Debug("usecase: getting bookings by user ID", logger.Int64("user_id", userID))

	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	bookings, err := uc.bookingRepo.GetBookingsByUserID(ctx, userID)
	if err != nil {
		logger.Error("usecase: failed to get bookings by user ID", logger.Int64("user_id", userID), logger.Err(err))
		return nil, err
	}

	logger.Debug("usecase: bookings fetched", logger.Int64("user_id", userID), logger.Int("count", len(bookings)))
	return bookings, nil
}

func (uc *bookingUsecase) GetAllBookings(ctx context.Context, status, sortBy, sortOrder string, page, limit int) ([]entity.BookingWithDetails, int, error) {
	logger.Debug("usecase: getting all bookings",
		logger.String("status", status),
		logger.Int("page", page),
		logger.Int("limit", limit),
	)

	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	bookings, total, err := uc.bookingRepo.GetAllBookings(ctx, status, sortBy, sortOrder, page, limit)
	if err != nil {
		logger.Error("usecase: failed to get all bookings", logger.Err(err))
		return nil, 0, err
	}

	logger.Debug("usecase: all bookings fetched", logger.Int("total", total))
	return bookings, total, nil
}

func (uc *bookingUsecase) GetBookingsByEventID(ctx context.Context, eventID int64, status, sortBy, sortOrder string) ([]entity.BookingWithDetails, error) {
	logger.Debug("usecase: getting bookings by event ID", logger.Int64("event_id", eventID))

	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	bookings, err := uc.bookingRepo.GetBookingsWithDetailsByEventID(ctx, eventID, status, sortBy, sortOrder)
	if err != nil {
		logger.Error("usecase: failed to get bookings by event ID", logger.Int64("event_id", eventID), logger.Err(err))
		return nil, err
	}

	logger.Debug("usecase: bookings fetched by event ID",
		logger.Int64("event_id", eventID),
		logger.Int("count", len(bookings)),
	)
	return bookings, nil
}