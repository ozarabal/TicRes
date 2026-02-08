package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ticres/internal/entity"
	"ticres/internal/repository"
	"ticres/pkg/logger"
)

type PaymentUsecase interface {
	ProcessPayment(ctx context.Context, bookingID, userID int64, paymentMethod string) (*entity.Transaction, error)
	GetPaymentStatus(ctx context.Context, bookingID, userID int64) (*entity.BookingWithPayment, error)
}

type paymentUsecase struct {
	bookingRepo     repository.BookingRepository
	transactionRepo repository.TransactionRepository
	contextTimeout  time.Duration
}

func NewPaymentUsecase(
	bookingRepo repository.BookingRepository,
	transactionRepo repository.TransactionRepository,
	timeout time.Duration,
) PaymentUsecase {
	return &paymentUsecase{
		bookingRepo:     bookingRepo,
		transactionRepo: transactionRepo,
		contextTimeout:  timeout,
	}
}

var validPaymentMethods = map[string]string{
	"credit_card":   "CR",
	"bank_transfer": "BT",
	"e_wallet":      "EW",
}

func (uc *paymentUsecase) ProcessPayment(ctx context.Context, bookingID, userID int64, paymentMethod string) (*entity.Transaction, error) {
	logger.Info("usecase: processing payment",
		logger.Int64("booking_id", bookingID),
		logger.Int64("user_id", userID),
		logger.String("payment_method", paymentMethod),
	)

	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	// Validate payment method
	methodCode, ok := validPaymentMethods[paymentMethod]
	if !ok {
		return nil, entity.ErrInvalidPaymentMethod
	}

	// Get booking and verify ownership
	booking, err := uc.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking.UserID != userID {
		return nil, entity.ErrUnauthorized
	}

	// Check booking status
	if booking.Status != "PENDING" {
		if booking.Status == "PAID" {
			return nil, entity.ErrPaymentAlreadyMade
		}
		return nil, entity.ErrBookingNotPending
	}

	// Check expiry
	if booking.ExpiresAt != nil && time.Now().After(*booking.ExpiresAt) {
		// Mark booking as expired and release seats
		uc.bookingRepo.UpdateBookingStatus(ctx, bookingID, "EXPIRED")
		uc.bookingRepo.ReleaseSeatsByBookingID(ctx, bookingID)
		return nil, entity.ErrBookingExpired
	}

	// Get or check existing transaction
	txn, err := uc.transactionRepo.GetTransactionByBookingID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if txn != nil && txn.Status == "COMPLETED" {
		return nil, entity.ErrPaymentAlreadyMade
	}

	// If no transaction exists yet, create one
	if txn == nil {
		txn = &entity.Transaction{
			Amount:        booking.TotalAmount,
			PaymentMethod: paymentMethod,
			BookingID:     bookingID,
			Status:        "PENDING",
		}
		if err := uc.transactionRepo.CreateTransaction(ctx, txn); err != nil {
			return nil, err
		}
	}

	// Simulate payment gateway processing
	time.Sleep(500 * time.Millisecond)

	// Generate external ID (mock gateway reference)
	externalID := fmt.Sprintf("PAY-%s-%d-%d", methodCode, bookingID, time.Now().UnixMilli())

	// Update transaction to COMPLETED
	if err := uc.transactionRepo.UpdateTransactionStatus(ctx, txn.ID, "COMPLETED", externalID); err != nil {
		logger.Error("usecase: failed to update transaction status", logger.Err(err))
		return nil, err
	}

	// Update booking to PAID
	if err := uc.bookingRepo.UpdateBookingStatus(ctx, bookingID, "PAID"); err != nil {
		logger.Error("usecase: failed to update booking status", logger.Err(err))
		return nil, err
	}

	txn.Status = "COMPLETED"
	txn.ExternalID = externalID
	txn.PaymentMethod = paymentMethod

	logger.Info("usecase: payment processed successfully",
		logger.Int64("booking_id", bookingID),
		logger.String("external_id", externalID),
		logger.String("payment_method", paymentMethod),
	)

	return txn, nil
}

func (uc *paymentUsecase) GetPaymentStatus(ctx context.Context, bookingID, userID int64) (*entity.BookingWithPayment, error) {
	logger.Debug("usecase: getting payment status", logger.Int64("booking_id", bookingID))

	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	booking, err := uc.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking.UserID != userID {
		return nil, entity.ErrUnauthorized
	}

	txn, err := uc.transactionRepo.GetTransactionByBookingID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	result := &entity.BookingWithPayment{
		BookingID:   booking.ID,
		EventID:     booking.EventID,
		Status:      booking.Status,
		TotalAmount: booking.TotalAmount,
		ExpiresAt:   booking.ExpiresAt,
		Transaction: txn,
	}

	return result, nil
}

// FormatPaymentMethod returns display name for a payment method code
func FormatPaymentMethod(method string) string {
	names := map[string]string{
		"credit_card":   "Credit Card",
		"bank_transfer": "Bank Transfer",
		"e_wallet":      "E-Wallet",
	}
	if name, ok := names[strings.ToLower(method)]; ok {
		return name
	}
	return method
}
