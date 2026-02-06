package repository

import (
	"context"
	"fmt"
	"ticres/internal/entity"
	"time"

	"encoding/json"

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
    // 1. Mulai Transaksi
    tx, err := r.db.Begin(ctx)
    if err != nil {
        return err
    }
    // Safety net: Rollback jika function selesai tanpa Commit
    defer tx.Rollback(ctx) 

    // 2. Insert Event
    queryEvent := `
        INSERT INTO events (name, location, date, capacity, created_at)
        VALUES ($1, $2, $3, $4, NOW())
        RETURNING event_id, created_at
    `
    // Perhatikan: Kita pakai `tx.QueryRow`, bukan `r.db.QueryRow`
    err = tx.QueryRow(ctx, queryEvent, event.Name, event.Location, event.Date, event.Capacity).Scan(&event.ID, &event.CreatedAt)
    if err != nil {
        return err
    }

    querySeat := `INSERT INTO seats (event_id, seat_number, is_booked) VALUES ($1, $2, False)`
    
    for i := 1; i <= event.Capacity; i++ {
        seatNum := fmt.Sprintf("%d-%d",event.ID , i)
        _, err := tx.Exec(ctx, querySeat, event.ID, seatNum)
        if err != nil {
            return err // Jika gagal generate kursi, Event juga batal dibuat (Rollback otomatis)
        }
    }

	r.redis.Del(ctx, eventsCacheKey)

    // 4. Commit Transaksi (Simpan Permanen)
    return tx.Commit(ctx)
}

func (r *eventRepository) GetAllEvents(ctx context.Context) ([]entity.Event, error){

	cachedData, err := r.redis.Get(ctx, eventsCacheKey).Result()
	if err == nil {
		var events []entity.Event
		if  err:= json.Unmarshal([]byte(cachedData), &events); err == nil {
			return events, nil
		}
	}

	query := `SELECT event_id ,name, location, date, capacity, created_at FROM events`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil , err
	}
	defer rows.Close()

	var events []entity.Event

	// scan setiap row, jadikan entity Event lalu masukan array
	for rows.Next(){
		var evt entity.Event

		err := rows.Scan(&evt.ID, &evt.Name, &evt.Location, &evt.Date, &evt.Capacity, &evt.CreatedAt)
		if err != nil {
			return nil, err
		}
		events = append(events,evt)
	}

	if data, err := json.Marshal(events); err == nil {
		r.redis.Set(ctx , eventsCacheKey, data, 10*time.Minute)
	}

	return events, nil 
}

func (r *eventRepository) GetEventByID(ctx context.Context, eventID int64) (*entity.Event, error) {
	// format unutk cacheKey
	key := fmt.Sprintf("events:detail:%d", eventID)
	var event entity.Event
	cachedData, err := r.redis.Get(ctx, key).Result()
	if err == nil {
		if  err:= json.Unmarshal([]byte(cachedData), &event); err == nil {
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
		return nil, err
	}

	return &event, nil

}

func (r *eventRepository) UpdateEvent(ctx context.Context, event *entity.Event, prevCapacity int64) error{
	// memulai transaksi
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	// rollback jika commit tidak terjadi
	defer tx.Rollback(ctx)

	//
	queryEvent := `
		UPDATE events
		SET name = $1, location = $2, date = $3, capacity = $4, updated_at = $5 
		WHERE event_id = $6
	`

	_ ,err = tx.Exec(ctx, queryEvent, event.Name, event.Location,event.Date, event.Capacity, event.UpdatedAt, event.ID)

	querySeats :=
	`
		INSERT INTO seats (event_id, seat_number, is_booked) VALUES ($1, $2, False)
	`

	for i := prevCapacity + 1; i <= int64(event.Capacity); i++{
		seatNum := fmt.Sprintf("%d-%d",event.ID , i)
		_, err = tx.Exec(ctx, querySeats, event.ID, seatNum)
		if err != nil{
			return err
		}
	}

	r.redis.Del(ctx, "events:list_all")

	return tx.Commit(ctx);
}

func (r *eventRepository) UpdateEventStatus(ctx context.Context, eventID int64, status string) error {
	query := `UPDATE events SET status = $1, updated_at = NOW() WHERE event_id = $2`
	_, err := r.db.Exec(ctx, query, status, eventID)

	r.redis.Del(ctx, "events:list_all")

	return err
}

func (r *eventRepository) GetEventsWithSearch(ctx context.Context, search string, page, limit int) ([]entity.Event, int, error) {
	// Count total
	countQuery := `SELECT COUNT(*) FROM events WHERE name ILIKE $1`
	searchPattern := "%" + search + "%"

	var total int
	err := r.db.QueryRow(ctx, countQuery, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated data
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
		return nil, 0, err
	}
	defer rows.Close()

	var events []entity.Event
	for rows.Next() {
		var evt entity.Event
		var status string
		err := rows.Scan(&evt.ID, &evt.Name, &evt.Location, &evt.Date, &evt.Capacity, &status, &evt.CreatedAt, &evt.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		events = append(events, evt)
	}

	return events, total, nil
}

func (r *eventRepository) GetEventWithSeats(ctx context.Context, eventID int64) (*entity.EventWithSeats, error) {
	// Get event
	event, err := r.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Get seats
	seats, err := r.GetSeatsByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	return &entity.EventWithSeats{
		Event: *event,
		Seats: seats,
	}, nil
}

func (r *eventRepository) GetSeatsByEventID(ctx context.Context, eventID int64) ([]entity.Seat, error) {
	query := `
		SELECT seat_id, event_id, seat_number, is_booked
		FROM seats
		WHERE event_id = $1
		ORDER BY seat_id
	`

	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seats []entity.Seat
	for rows.Next() {
		var seat entity.Seat
		err := rows.Scan(&seat.ID, &seat.EventID, &seat.SeatNumber, &seat.IsBooked)
		if err != nil {
			return nil, err
		}
		seats = append(seats, seat)
	}

	return seats, nil
}