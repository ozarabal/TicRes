package repository

import (
	"context"
	"errors"
	"fmt"

	"ticres/internal/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingRepository interface {
	CreateBooking(ctx context.Context, userID, eventID int64, seatIDs []int64) (int64, error)
	GetBookingsByEventID(ctx context.Context, eventID int64) ([]entity.Booking, error)
	GetBookingsByUserID(ctx context.Context, userID int64) ([]entity.BookingWithDetails, error)
	GetAllBookings(ctx context.Context, status, sortBy, sortOrder string, page, limit int) ([]entity.BookingWithDetails, int, error)
	GetBookingsWithDetailsByEventID(ctx context.Context, eventID int64, status, sortBy, sortOrder string) ([]entity.BookingWithDetails, error)
	UpdateBookingStatus(ctx context.Context, bookingID int64, status string) error
}

type bookingRepository struct {
	db *pgxpool.Pool
}

func NewBookingRepository(db *pgxpool.Pool) BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) CreateBooking(ctx context.Context, userID, eventID int64, seatIDs []int64) (int64, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var bookingID int64
	queryBooking := `
		INSERT INTO booking (user_id, event_id, status, created_at)
		VALUES ($1, $2, 'PENDING', NOW())
		RETURNING booking_id
	`
	err = tx.QueryRow(ctx, queryBooking, userID, eventID).Scan(&bookingID)
	if err != nil {
		return 0, err
	}

	queryLockSeat := `UPDATE seats SET is_booked = True WHERE seat_id = $1 AND is_booked = False`
	queryInsertItem := `INSERT INTO booking_items (booking_id, seat_id) VALUES ($1, $2)`

	for _, seatID := range seatIDs {
		cmdTag, err := tx.Exec(ctx, queryLockSeat, seatID)
		if err != nil {
			return 0, err
		}
		if cmdTag.RowsAffected() == 0 {
			return 0, errors.New("seat not available or already booked")
		}
		_, err = tx.Exec(ctx, queryInsertItem, bookingID, seatID)
		if err != nil {
			return 0, err
		}
	}
	return bookingID, tx.Commit(ctx)
}

func (r *bookingRepository) GetBookingsByEventID(ctx context.Context, eventID int64) ([]entity.Booking, error) {
	query := `
		SELECT booking_id, user_id, event_id, status, created_at
		FROM booking
		WHERE event_id = $1 AND status IN ('PAID', 'PENDING')
	`
	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []entity.Booking
	for rows.Next() {
		var b entity.Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.EventID, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, nil
}

func (r *bookingRepository) GetBookingsByUserID(ctx context.Context, userID int64) ([]entity.BookingWithDetails, error) {
	query := `
		SELECT b.booking_id, b.user_id, u.name, u.email, b.event_id, e.name, b.status, b.created_at
		FROM booking b
		JOIN users u ON b.user_id = u.user_id
		JOIN events e ON b.event_id = e.event_id
		WHERE b.user_id = $1
		ORDER BY b.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []entity.BookingWithDetails
	for rows.Next() {
		var b entity.BookingWithDetails
		if err := rows.Scan(&b.ID, &b.UserID, &b.UserName, &b.UserEmail, &b.EventID, &b.EventName, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, nil
}

func (r *bookingRepository) GetAllBookings(ctx context.Context, status, sortBy, sortOrder string, page, limit int) ([]entity.BookingWithDetails, int, error) {
	// Build query with filters
	baseQuery := `
		FROM booking b
		JOIN users u ON b.user_id = u.user_id
		JOIN events e ON b.event_id = e.event_id
	`
	whereClause := ""
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		whereClause = fmt.Sprintf(" WHERE b.status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	// Count total
	countQuery := "SELECT COUNT(*) " + baseQuery + whereClause
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Validate sort fields
	validSortFields := map[string]string{
		"created_at": "b.created_at",
		"status":     "b.status",
	}
	sortField, ok := validSortFields[sortBy]
	if !ok {
		sortField = "b.created_at"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Build data query
	offset := (page - 1) * limit
	dataQuery := fmt.Sprintf(`
		SELECT b.booking_id, b.user_id, u.name, u.email, b.event_id, e.name, b.status, b.created_at
		%s%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, baseQuery, whereClause, sortField, sortOrder, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var bookings []entity.BookingWithDetails
	for rows.Next() {
		var b entity.BookingWithDetails
		if err := rows.Scan(&b.ID, &b.UserID, &b.UserName, &b.UserEmail, &b.EventID, &b.EventName, &b.Status, &b.CreatedAt); err != nil {
			return nil, 0, err
		}
		bookings = append(bookings, b)
	}
	return bookings, total, nil
}

func (r *bookingRepository) GetBookingsWithDetailsByEventID(ctx context.Context, eventID int64, status, sortBy, sortOrder string) ([]entity.BookingWithDetails, error) {
	baseQuery := `
		SELECT b.booking_id, b.user_id, u.name, u.email, b.event_id, e.name, b.status, b.created_at
		FROM booking b
		JOIN users u ON b.user_id = u.user_id
		JOIN events e ON b.event_id = e.event_id
		WHERE b.event_id = $1
	`
	args := []interface{}{eventID}
	argIndex := 2

	if status != "" {
		baseQuery += fmt.Sprintf(" AND b.status = $%d", argIndex)
		args = append(args, status)
	}

	// Validate sort fields
	validSortFields := map[string]string{
		"created_at": "b.created_at",
		"status":     "b.status",
	}
	sortField, ok := validSortFields[sortBy]
	if !ok {
		sortField = "b.created_at"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortField, sortOrder)

	rows, err := r.db.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []entity.BookingWithDetails
	for rows.Next() {
		var b entity.BookingWithDetails
		if err := rows.Scan(&b.ID, &b.UserID, &b.UserName, &b.UserEmail, &b.EventID, &b.EventName, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, nil
}

func (r *bookingRepository) UpdateBookingStatus(ctx context.Context, bookingID int64, status string) error {
	query := `UPDATE booking SET status = $1 WHERE booking_id = $2`
	_, err := r.db.Exec(ctx, query, status, bookingID)
	return err
}
