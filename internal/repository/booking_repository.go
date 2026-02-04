package repository

import (
	"context"
	"errors"

	"ticres/internal/entity"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingRepository interface {
	CreateBooking(ctx context.Context, userID, eventID int64, seatIDs []int64) (int64, error)
	GetBookingsByEventID(ctx context.Context, eventID int64) ([]entity.Booking, error)
	UpdateBookingStatus(ctx context.Context, bookingID int64, status string) error 
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

func (r *bookingRepository) GetBookingsByEventID(ctx context.Context, eventID int64) ([]entity.Booking, error) {
    query := `
        SELECT booking_id, user_id, event_id, status, total_amount, email -- Asumsi ada kolom email di tabel bookings atau join ke users
        FROM bookings 
        WHERE event_id = $1 AND status IN ('PAID', 'PENDING')
    `
    // Note: Jika email ada di tabel users, Anda harus melakukan JOIN SQL disini.
    // Contoh: SELECT b.booking_id, u.email ... FROM bookings b JOIN users u ON b.user_id = u.user_id ...
    
    rows, err := r.db.Query(ctx, query, eventID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var bookings []entity.Booking
    for rows.Next() {
        var b entity.Booking
        // Scan sesuai struktur entity Anda
        if err := rows.Scan(&b.ID, &b.UserID, &b.EventID, &b.Status); err != nil {
            return nil, err
        }
        bookings = append(bookings, b)
    }
    return bookings, nil
}

func (r *bookingRepository) UpdateBookingStatus(ctx context.Context, bookingID int64, status string) error {
    query := `UPDATE bookings SET status = $1 WHERE booking_id = $2`
    _, err := r.db.Exec(ctx, query, status, bookingID)
    return err
}

