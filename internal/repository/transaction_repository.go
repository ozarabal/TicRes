package repository

import (
	"context"
	"fmt"
	"time"

	"ticres/internal/entity"
	"ticres/pkg/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, txn *entity.Transaction) error
	GetTransactionByBookingID(ctx context.Context, bookingID int64) (*entity.Transaction, error)
	GetTransactionByExternalID(ctx context.Context, externalID string) (*entity.Transaction, error)
	UpdateTransactionStatus(ctx context.Context, paymentID int64, status, externalID string) error
}

type transactionRepository struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) CreateTransaction(ctx context.Context, txn *entity.Transaction) error {
	logger.Debug("creating transaction",
		logger.Int64("booking_id", txn.BookingID),
		logger.Float64("amount", txn.Amount),
	)

	query := `
		INSERT INTO transactions (amount, payment_method, booking_id, external_id, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING payment_id, transaction_date
	`

	externalID := fmt.Sprintf("TXN-%d-%d", txn.BookingID, time.Now().UnixMilli())

	err := r.db.QueryRow(ctx, query,
		txn.Amount, txn.PaymentMethod, txn.BookingID, externalID, "PENDING",
	).Scan(&txn.ID, &txn.TransactionDate)
	if err != nil {
		logger.Error("failed to create transaction", logger.Err(err))
		return err
	}

	txn.ExternalID = externalID
	txn.Status = "PENDING"

	logger.Info("transaction created",
		logger.Int64("payment_id", txn.ID),
		logger.Int64("booking_id", txn.BookingID),
		logger.String("external_id", externalID),
	)
	return nil
}

func (r *transactionRepository) GetTransactionByBookingID(ctx context.Context, bookingID int64) (*entity.Transaction, error) {
	logger.Debug("fetching transaction by booking ID", logger.Int64("booking_id", bookingID))

	query := `
		SELECT payment_id, amount, COALESCE(payment_method, ''), booking_id, transaction_date, COALESCE(external_id, ''), COALESCE(status, 'PENDING')
		FROM transactions
		WHERE booking_id = $1
	`

	var txn entity.Transaction
	err := r.db.QueryRow(ctx, query, bookingID).Scan(
		&txn.ID, &txn.Amount, &txn.PaymentMethod, &txn.BookingID,
		&txn.TransactionDate, &txn.ExternalID, &txn.Status,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		logger.Error("failed to fetch transaction", logger.Int64("booking_id", bookingID), logger.Err(err))
		return nil, err
	}

	return &txn, nil
}

func (r *transactionRepository) GetTransactionByExternalID(ctx context.Context, externalID string) (*entity.Transaction, error) {
	logger.Debug("fetching transaction by external ID", logger.String("external_id", externalID))

	query := `
		SELECT payment_id, amount, COALESCE(payment_method, ''), booking_id, transaction_date, COALESCE(external_id, ''), COALESCE(status, 'PENDING')
		FROM transactions
		WHERE external_id = $1
	`

	var txn entity.Transaction
	err := r.db.QueryRow(ctx, query, externalID).Scan(
		&txn.ID, &txn.Amount, &txn.PaymentMethod, &txn.BookingID,
		&txn.TransactionDate, &txn.ExternalID, &txn.Status,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		logger.Error("failed to fetch transaction by external ID", logger.String("external_id", externalID), logger.Err(err))
		return nil, err
	}

	return &txn, nil
}

func (r *transactionRepository) UpdateTransactionStatus(ctx context.Context, paymentID int64, status, externalID string) error {
	logger.Debug("updating transaction status",
		logger.Int64("payment_id", paymentID),
		logger.String("status", status),
	)

	query := `UPDATE transactions SET status = $1, payment_method = COALESCE(NULLIF($2, ''), payment_method), external_id = COALESCE(NULLIF($3, ''), external_id) WHERE payment_id = $4`
	_, err := r.db.Exec(ctx, query, status, "", externalID, paymentID)
	if err != nil {
		logger.Error("failed to update transaction status",
			logger.Int64("payment_id", paymentID),
			logger.Err(err),
		)
		return err
	}

	logger.Info("transaction status updated",
		logger.Int64("payment_id", paymentID),
		logger.String("status", status),
	)
	return nil
}
