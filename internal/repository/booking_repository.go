package repository

import (
	"context"
	"errors"
	"fmt"

	"ticres/internal/entity"
	"ticres/pkg/logger"

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
	logger.Debug("creating booking",
		logger.Int64("user_id", userID),
		logger.Int64("event_id", eventID),
		logger.Int("seat_count", len(seatIDs)),
	)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		logger.Error("failed to begin transaction", logger.Err(err))
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
		logger.Error("failed to insert booking", logger.Err(err))
		return 0, err
	}

	queryLockSeat := `UPDATE seats SET is_booked = True WHERE seat_id = $1 AND is_booked = False`
	queryInsertItem := `INSERT INTO booking_items (booking_id, seat_id) VALUES ($1, $2)`

	for _, seatID := range seatIDs {
		cmdTag, err := tx.Exec(ctx, queryLockSeat, seatID)
		if err != nil {
			logger.Error("failed to lock seat",
				logger.Int64("seat_id", seatID),
				logger.Err(err),
			)
			return 0, err
		}
		if cmdTag.RowsAffected() == 0 {
			logger.Warn("seat not available",
				logger.Int64("seat_id", seatID),
				logger.Int64("booking_id", bookingID),
			)
			return 0, errors.New("seat not available or already booked")
		}
		_, err = tx.Exec(ctx, queryInsertItem, bookingID, seatID)
		if err != nil {
			logger.Error("failed to insert booking item", logger.Err(err))
			return 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		logger.Error("failed to commit booking transaction", logger.Err(err))
		return 0, err
	}

	logger.Info("booking created successfully",
		logger.Int64("booking_id", bookingID),
		logger.Int64("user_id", userID),
		logger.Int64("event_id", eventID),
		logger.Int("seat_count", len(seatIDs)),
	)
	return bookingID, nil
}

func (r *bookingRepository) GetBookingsByEventID(ctx context.Context, eventID int64) ([]entity.Booking, error) {
	logger.Debug("fetching bookings by event ID", logger.Int64("event_id", eventID))

	query := `
		SELECT booking_id, user_id, event_id, status, created_at
		FROM booking
		WHERE event_id = $1 AND status IN ('PAID', 'PENDING')
	`
	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		logger.Error("failed to query bookings by event ID", logger.Int64("event_id", eventID), logger.Err(err))
		return nil, err
	}
	defer rows.Close()

	var bookings []entity.Booking
	for rows.Next() {
		var b entity.Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.EventID, &b.Status, &b.CreatedAt); err != nil {
			logger.Error("failed to scan booking row", logger.Err(err))
			return nil, err
		}
		bookings = append(bookings, b)
	}

	logger.Debug("bookings fetched by event ID",
		logger.Int64("event_id", eventID),
		logger.Int("count", len(bookings)),
	)
	return bookings, nil
}

func (r *bookingRepository) GetBookingsByUserID(ctx context.Context, userID int64) ([]entity.BookingWithDetails, error) {
	logger.Debug("fetching bookings by user ID", logger.Int64("user_id", userID))

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
		logger.Error("failed to query bookings by user ID", logger.Int64("user_id", userID), logger.Err(err))
		return nil, err
	}
	defer rows.Close()

	var bookings []entity.BookingWithDetails
	for rows.Next() {
		var b entity.BookingWithDetails
		if err := rows.Scan(&b.ID, &b.UserID, &b.UserName, &b.UserEmail, &b.EventID, &b.EventName, &b.Status, &b.CreatedAt); err != nil {
			logger.Error("failed to scan booking row", logger.Err(err))
			return nil, err
		}
		bookings = append(bookings, b)
	}

	logger.Debug("bookings fetched by user ID",
		logger.Int64("user_id", userID),
		logger.Int("count", len(bookings)),
	)
	return bookings, nil
}

func (r *bookingRepository) GetAllBookings(ctx context.Context, status, sortBy, sortOrder string, page, limit int) ([]entity.BookingWithDetails, int, error) {
	logger.Debug("fetching all bookings",
		logger.String("status", status),
		logger.String("sort_by", sortBy),
		logger.String("sort_order", sortOrder),
		logger.Int("page", page),
		logger.Int("limit", limit),
	)

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

	countQuery := "SELECT COUNT(*) " + baseQuery + whereClause
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		logger.Error("failed to count bookings", logger.Err(err))
		return nil, 0, err
	}

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
		logger.Error("failed to query all bookings", logger.Err(err))
		return nil, 0, err
	}
	defer rows.Close()

	var bookings []entity.BookingWithDetails
	for rows.Next() {
		var b entity.BookingWithDetails
		if err := rows.Scan(&b.ID, &b.UserID, &b.UserName, &b.UserEmail, &b.EventID, &b.EventName, &b.Status, &b.CreatedAt); err != nil {
			logger.Error("failed to scan booking row", logger.Err(err))
			return nil, 0, err
		}
		bookings = append(bookings, b)
	}

	logger.Debug("all bookings fetched",
		logger.Int("total", total),
		logger.Int("returned", len(bookings)),
	)
	return bookings, total, nil
}

func (r *bookingRepository) GetBookingsWithDetailsByEventID(ctx context.Context, eventID int64, status, sortBy, sortOrder string) ([]entity.BookingWithDetails, error) {
	logger.Debug("fetching bookings with details by event ID",
		logger.Int64("event_id", eventID),
		logger.String("status", status),
		logger.String("sort_by", sortBy),
		logger.String("sort_order", sortOrder),
	)

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
		logger.Error("failed to query bookings with details by event ID",
			logger.Int64("event_id", eventID),
			logger.Err(err),
		)
		return nil, err
	}
	defer rows.Close()

	var bookings []entity.BookingWithDetails
	for rows.Next() {
		var b entity.BookingWithDetails
		if err := rows.Scan(&b.ID, &b.UserID, &b.UserName, &b.UserEmail, &b.EventID, &b.EventName, &b.Status, &b.CreatedAt); err != nil {
			logger.Error("failed to scan booking row", logger.Err(err))
			return nil, err
		}
		bookings = append(bookings, b)
	}

	logger.Debug("bookings with details fetched by event ID",
		logger.Int64("event_id", eventID),
		logger.Int("count", len(bookings)),
	)
	return bookings, nil
}

func (r *bookingRepository) UpdateBookingStatus(ctx context.Context, bookingID int64, status string) error {
	logger.Debug("updating booking status",
		logger.Int64("booking_id", bookingID),
		logger.String("status", status),
	)

	query := `UPDATE booking SET status = $1 WHERE booking_id = $2`
	_, err := r.db.Exec(ctx, query, status, bookingID)
	if err != nil {
		logger.Error("failed to update booking status",
			logger.Int64("booking_id", bookingID),
			logger.String("status", status),
			logger.Err(err),
		)
		return err
	}

	logger.Info("booking status updated",
		logger.Int64("booking_id", bookingID),
		logger.String("status", status),
	)
	return nil
}
