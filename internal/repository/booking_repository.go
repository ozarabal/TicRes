package repository

import (
	"context"
	"errors"

	// "ticres/internal/entity"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingRepository interface {
	CreateBooking(ctx context.Context, userID, eventID int64, seatIDs []int64) (int64, error)
}

type bookingRepository struct {
	db *pgxpool.Pool
}

func NewBookingRepository(db *pgxpool.Pool) BookingRepository{
	return &bookingRepository{db:db}
}

func (r *bookingRepository) CreateBooking(ctx context.Context,userID, eventID int64, seatIDs []int64) (int64 ,error) {

	// begin transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0 ,err
	}

	defer tx.Rollback(ctx)

	// insert Header Booking
	var bookingID int64

	queryBooking := `
		INSERT INTO booking (user_id, event_id, status, created_at) VALUES
		($1,$2,'PENDING', NOW()) RETURNING booking_id
	`

	err = tx.QueryRow(ctx, queryBooking, userID, eventID).Scan(&bookingID)
	if err != nil {
		return 0 ,err
	}

	// locking seat

	queryLockSeat := `
		UPDATE seats
		SET is_booked = True
		WHERE seat_id = $1 AND is_booked = False
	`
	queryInsertItem := `
		INSERT INTO booking_items (booking_id, seat_id)
		VALUES ($1, $2)
	`

	for _, seatID := range seatIDs {

		cmdTag, err := tx.Exec(ctx, queryLockSeat, seatID)
		if err != nil {
			return 0 , err
		}

		if cmdTag.RowsAffected() == 0 {
			return 0 ,errors.New("seat not available or already booked")
		}

		_, err = tx.Exec(ctx, queryInsertItem, bookingID, seatID)
		if err != nil {
			return 0 , err
		}
	}
	return bookingID ,tx.Commit(ctx)
}

