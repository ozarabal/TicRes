package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ticres/internal/entity"
	"ticres/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type EventRepository interface {
	CreateEvent(ctx context.Context, event *entity.Event) error
	GetAllEvents(ctx context.Context) ([]entity.Event, error)
	GetEventsWithSearch(ctx context.Context, search string, page, limit int) ([]entity.Event, int, error)
	GetEventByID(ctx context.Context, eventID int64) (*entity.Event, error)
	GetEventWithSeats(ctx context.Context, eventID int64) (*entity.EventWithSeats, error)
	GetSeatsByEventID(ctx context.Context, eventID int64) ([]entity.Seat, error)
	UpdateEvent(ctx context.Context, event *entity.Event, preCapacity int64) error
	UpdateEventStatus(ctx context.Context, eventID int64, status string) error
}

type eventRepository struct {
	db *pgxpool.Pool
	redis *redis.Client
}

func NewEventRepository(db *pgxpool.Pool, rdb *redis.Client) EventRepository {
	return &eventRepository{db:db, redis:rdb}
}

const eventsCacheKey = "events:list_all"

func (r *eventRepository) CreateEvent(ctx context.Context, event *entity.Event) error {
	logger.Debug("creating event",
		logger.String("name", event.Name),
		logger.String("location", event.Location),
		logger.Int("capacity", event.Capacity),
	)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		logger.Error("failed to begin transaction", logger.Err(err))
		return err
	}
	defer tx.Rollback(ctx)

	queryEvent := `
		INSERT INTO events (name, location, date, capacity, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING event_id, created_at
	`
	err = tx.QueryRow(ctx, queryEvent, event.Name, event.Location, event.Date, event.Capacity).Scan(&event.ID, &event.CreatedAt)
	if err != nil {
		logger.Error("failed to insert event", logger.Err(err))
		return err
	}

	querySeat := `INSERT INTO seats (event_id, seat_number, is_booked) VALUES ($1, $2, False)`

	for i := 1; i <= event.Capacity; i++ {
		seatNum := fmt.Sprintf("%d-%d", event.ID, i)
		_, err := tx.Exec(ctx, querySeat, event.ID, seatNum)
		if err != nil {
			logger.Error("failed to create seat",
				logger.Int64("event_id", event.ID),
				logger.Int("seat_number", i),
				logger.Err(err),
			)
			return err
		}
	}

	r.redis.Del(ctx, eventsCacheKey)

	if err := tx.Commit(ctx); err != nil {
		logger.Error("failed to commit transaction", logger.Err(err))
		return err
	}

	logger.Info("event created successfully",
		logger.Int64("event_id", event.ID),
		logger.String("name", event.Name),
		logger.Int("capacity", event.Capacity),
	)
	return nil
}

func (r *eventRepository) GetAllEvents(ctx context.Context) ([]entity.Event, error) {
	logger.Debug("fetching all events")

	cachedData, err := r.redis.Get(ctx, eventsCacheKey).Result()
	if err == nil {
		var events []entity.Event
		if err := json.Unmarshal([]byte(cachedData), &events); err == nil {
			logger.Debug("events fetched from cache", logger.Int("count", len(events)))
			return events, nil
		}
	}

	query := `SELECT event_id ,name, location, date, capacity, created_at FROM events`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		logger.Error("failed to query events", logger.Err(err))
		return nil, err
	}
	defer rows.Close()

	var events []entity.Event
	for rows.Next() {
		var evt entity.Event
		err := rows.Scan(&evt.ID, &evt.Name, &evt.Location, &evt.Date, &evt.Capacity, &evt.CreatedAt)
		if err != nil {
			logger.Error("failed to scan event row", logger.Err(err))
			return nil, err
		}
		events = append(events, evt)
	}

	if data, err := json.Marshal(events); err == nil {
		r.redis.Set(ctx, eventsCacheKey, data, 10*time.Minute)
		logger.Debug("events cached", logger.Int("count", len(events)))
	}

	logger.Debug("events fetched from database", logger.Int("count", len(events)))
	return events, nil
}

func (r *eventRepository) GetEventByID(ctx context.Context, eventID int64) (*entity.Event, error) {
	logger.Debug("fetching event by ID", logger.Int64("event_id", eventID))

	key := fmt.Sprintf("events:detail:%d", eventID)
	var event entity.Event
	cachedData, err := r.redis.Get(ctx, key).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(cachedData), &event); err == nil {
			logger.Debug("event fetched from cache", logger.Int64("event_id", eventID))
			return &event, nil
		}
	}

	query := `SELECT event_id ,name, location, date, capacity, created_at FROM events WHERE event_id=$1`

	err = r.db.QueryRow(ctx, query, eventID).Scan(
		&event.ID,
		&event.Name,
		&event.Location,
		&event.Date,
		&event.Capacity,
		&event.CreatedAt,
	)

	if err != nil {
		logger.Warn("event not found", logger.Int64("event_id", eventID), logger.Err(err))
		return nil, err
	}

	logger.Debug("event fetched from database", logger.Int64("event_id", eventID))
	return &event, nil
}

func (r *eventRepository) UpdateEvent(ctx context.Context, event *entity.Event, prevCapacity int64) error {
	logger.Debug("updating event",
		logger.Int64("event_id", event.ID),
		logger.String("name", event.Name),
		logger.Int64("prev_capacity", prevCapacity),
		logger.Int("new_capacity", event.Capacity),
	)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		logger.Error("failed to begin transaction", logger.Err(err))
		return err
	}
	defer tx.Rollback(ctx)

	queryEvent := `
		UPDATE events
		SET name = $1, location = $2, date = $3, capacity = $4, updated_at = $5
		WHERE event_id = $6
	`

	_, err = tx.Exec(ctx, queryEvent, event.Name, event.Location, event.Date, event.Capacity, event.UpdatedAt, event.ID)
	if err != nil {
		logger.Error("failed to update event", logger.Int64("event_id", event.ID), logger.Err(err))
		return err
	}

	querySeats := `INSERT INTO seats (event_id, seat_number, is_booked) VALUES ($1, $2, False)`

	for i := prevCapacity + 1; i <= int64(event.Capacity); i++ {
		seatNum := fmt.Sprintf("%d-%d", event.ID, i)
		_, err = tx.Exec(ctx, querySeats, event.ID, seatNum)
		if err != nil {
			logger.Error("failed to create new seat",
				logger.Int64("event_id", event.ID),
				logger.Int64("seat_number", i),
				logger.Err(err),
			)
			return err
		}
	}

	r.redis.Del(ctx, "events:list_all")

	if err := tx.Commit(ctx); err != nil {
		logger.Error("failed to commit transaction", logger.Err(err))
		return err
	}

	logger.Info("event updated successfully", logger.Int64("event_id", event.ID))
	return nil
}

func (r *eventRepository) UpdateEventStatus(ctx context.Context, eventID int64, status string) error {
	logger.Debug("updating event status",
		logger.Int64("event_id", eventID),
		logger.String("status", status),
	)

	query := `UPDATE events SET status = $1, updated_at = NOW() WHERE event_id = $2`
	_, err := r.db.Exec(ctx, query, status, eventID)
	if err != nil {
		logger.Error("failed to update event status",
			logger.Int64("event_id", eventID),
			logger.String("status", status),
			logger.Err(err),
		)
		return err
	}

	r.redis.Del(ctx, "events:list_all")

	logger.Info("event status updated",
		logger.Int64("event_id", eventID),
		logger.String("status", status),
	)
	return nil
}

func (r *eventRepository) GetEventsWithSearch(ctx context.Context, search string, page, limit int) ([]entity.Event, int, error) {
	logger.Debug("searching events",
		logger.String("search", search),
		logger.Int("page", page),
		logger.Int("limit", limit),
	)

	countQuery := `SELECT COUNT(*) FROM events WHERE name ILIKE $1`
	searchPattern := "%" + search + "%"

	var total int
	err := r.db.QueryRow(ctx, countQuery, searchPattern).Scan(&total)
	if err != nil {
		logger.Error("failed to count events", logger.Err(err))
		return nil, 0, err
	}

	offset := (page - 1) * limit
	query := `
		SELECT event_id, name, location, date, capacity, COALESCE(status, 'available') as status, created_at, COALESCE(updated_at, created_at) as updated_at
		FROM events
		WHERE name ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, searchPattern, limit, offset)
	if err != nil {
		logger.Error("failed to query events with search", logger.Err(err))
		return nil, 0, err
	}
	defer rows.Close()

	var events []entity.Event
	for rows.Next() {
		var evt entity.Event
		var status string
		err := rows.Scan(&evt.ID, &evt.Name, &evt.Location, &evt.Date, &evt.Capacity, &status, &evt.CreatedAt, &evt.UpdatedAt)
		if err != nil {
			logger.Error("failed to scan event row", logger.Err(err))
			return nil, 0, err
		}
		events = append(events, evt)
	}

	logger.Debug("events search completed",
		logger.String("search", search),
		logger.Int("total", total),
		logger.Int("returned", len(events)),
	)
	return events, total, nil
}

func (r *eventRepository) GetEventWithSeats(ctx context.Context, eventID int64) (*entity.EventWithSeats, error) {
	logger.Debug("fetching event with seats", logger.Int64("event_id", eventID))

	event, err := r.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	seats, err := r.GetSeatsByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	logger.Debug("event with seats fetched",
		logger.Int64("event_id", eventID),
		logger.Int("seat_count", len(seats)),
	)
	return &entity.EventWithSeats{
		Event: *event,
		Seats: seats,
	}, nil
}

func (r *eventRepository) GetSeatsByEventID(ctx context.Context, eventID int64) ([]entity.Seat, error) {
	logger.Debug("fetching seats by event ID", logger.Int64("event_id", eventID))

	query := `
		SELECT seat_id, event_id, seat_number, is_booked
		FROM seats
		WHERE event_id = $1
		ORDER BY seat_id
	`

	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		logger.Error("failed to query seats", logger.Int64("event_id", eventID), logger.Err(err))
		return nil, err
	}
	defer rows.Close()

	var seats []entity.Seat
	for rows.Next() {
		var seat entity.Seat
		err := rows.Scan(&seat.ID, &seat.EventID, &seat.SeatNumber, &seat.IsBooked)
		if err != nil {
			logger.Error("failed to scan seat row", logger.Err(err))
			return nil, err
		}
		seats = append(seats, seat)
	}

	logger.Debug("seats fetched", logger.Int64("event_id", eventID), logger.Int("count", len(seats)))
	return seats, nil
}