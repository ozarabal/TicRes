package usecase

import (
	"context"
	"time"

	"ticres/internal/entity"
	"ticres/internal/repository"
	"ticres/pkg/logger"
)

type BookingUsecase interface {
	BookSeats(ctx context.Context, userID, eventID int64, seatIDs []int64, userEmail string) (*entity.BookingWithPayment, error)
	GetBookingsByUserID(ctx context.Context, userID int64) ([]entity.BookingWithDetails, error)
	GetAllBookings(ctx context.Context, status, sortBy, sortOrder string, page, limit int) ([]entity.BookingWithDetails, int, error)
	GetBookingsByEventID(ctx context.Context, eventID int64, status, sortBy, sortOrder string) ([]entity.BookingWithDetails, error)
}

type NotificationService interface {
	SendNotification(bookingID int64, email, message string)
	EnqueueCancellation(eventID int64)
}

type bookingUsecase struct {
	bookingRepo     repository.BookingRepository
	transactionRepo repository.TransactionRepository
	contextTimeout  time.Duration
	notifWorker     NotificationService
}

func NewBookingUsecase(repo repository.BookingRepository, txnRepo repository.TransactionRepository, timeout time.Duration, notifWorker NotificationService) BookingUsecase {
	return &bookingUsecase{
		bookingRepo:     repo,
		transactionRepo: txnRepo,
		contextTimeout:  timeout,
		notifWorker:     notifWorker,
	}
}

func (uc *bookingUsecase) BookSeats(ctx context.Context, userID, eventID int64, seatIDs []int64, userEmail string) (*entity.BookingWithPayment, error) {
	logger.Debug("usecase: booking seats",
		logger.Int64("user_id", userID),
		logger.Int64("event_id", eventID),
		logger.Int("seat_count", len(seatIDs)),
	)

	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	bookingID, totalAmount, err := uc.bookingRepo.CreateBooking(ctx, userID, eventID, seatIDs)
	if err != nil {
		logger.Error("usecase: failed to book seats",
			logger.Int64("user_id", userID),
			logger.Int64("event_id", eventID),
			logger.Err(err),
		)
		return nil, err
	}

	// Create a PENDING transaction
	txn := &entity.Transaction{
		Amount:    totalAmount,
		BookingID: bookingID,
		Status:    "PENDING",
	}
	if err := uc.transactionRepo.CreateTransaction(ctx, txn); err != nil {
		logger.Error("usecase: failed to create pending transaction",
			logger.Int64("booking_id", bookingID),
			logger.Err(err),
		)
		// Booking was created successfully, so we don't fail the whole operation
		// The transaction can be created later during payment
	}

	expiresAt := time.Now().Add(15 * time.Minute)
	uc.notifWorker.SendNotification(bookingID, userEmail, "Booking berhasil! Silakan selesaikan pembayaran dalam 15 menit.")

	logger.Info("usecase: seats booked successfully",
		logger.Int64("booking_id", bookingID),
		logger.Int64("user_id", userID),
		logger.Int64("event_id", eventID),
		logger.Float64("total_amount", totalAmount),
	)

	return &entity.BookingWithPayment{
		BookingID:   bookingID,
		EventID:     eventID,
		Status:      "PENDING",
		TotalAmount: totalAmount,
		ExpiresAt:   &expiresAt,
		Transaction: txn,
	}, nil
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
