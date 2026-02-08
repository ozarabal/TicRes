package repository

import (
	"context"

	"ticres/internal/entity"
	"ticres/pkg/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefundRepository interface {
	CreateRefund(ctx context.Context, refund *entity.Refund) error
	GetRefundByBookingID(ctx context.Context, bookingID int64) (*entity.Refund, error)
}

type refundRepository struct {
	db *pgxpool.Pool
}

func NewRefundRepository(db *pgxpool.Pool) RefundRepository {
	return &refundRepository{db: db}
}

func (r *refundRepository) CreateRefund(ctx context.Context, refund *entity.Refund) error {
	logger.Debug("creating refund",
		logger.Int64("booking_id", refund.BookingID),
		logger.Float64("amount", refund.Amount),
		logger.String("reason", refund.Reason),
	)

	query := `
		INSERT INTO refund (booking_id, amount, reason, status)
		VALUES ($1, $2, $3, $4)
		RETURNING refund_id, refund_date
	`

	err := r.db.QueryRow(ctx, query,
		refund.BookingID, refund.Amount, refund.Reason, "COMPLETED",
	).Scan(&refund.ID, &refund.RefundDate)
	if err != nil {
		logger.Error("failed to create refund", logger.Err(err))
		return err
	}

	refund.Status = "COMPLETED"

	logger.Info("refund created",
		logger.Int64("refund_id", refund.ID),
		logger.Int64("booking_id", refund.BookingID),
		logger.Float64("amount", refund.Amount),
	)
	return nil
}

func (r *refundRepository) GetRefundByBookingID(ctx context.Context, bookingID int64) (*entity.Refund, error) {
	logger.Debug("fetching refund by booking ID", logger.Int64("booking_id", bookingID))

	query := `
		SELECT refund_id, booking_id, amount, refund_date, COALESCE(reason, ''), COALESCE(status, 'PENDING')
		FROM refund
		WHERE booking_id = $1
	`

	var refund entity.Refund
	err := r.db.QueryRow(ctx, query, bookingID).Scan(
		&refund.ID, &refund.BookingID, &refund.Amount,
		&refund.RefundDate, &refund.Reason, &refund.Status,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		logger.Error("failed to fetch refund", logger.Int64("booking_id", bookingID), logger.Err(err))
		return nil, err
	}

	return &refund, nil
}
