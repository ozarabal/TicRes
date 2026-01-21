package repository

import(

	"context"
	"ticres/internal/entity"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository interface{
	CreateEvent(ctx context.Context, event *entity.Event) error
	GetAllEvents(ctx context.Context) ([]entity.Event, error)
}

type eventRepository struct {
	db *pgxpool.Pool
}

func NewEventRepository(db *pgxpool.Pool) EventRepository {
	return &eventRepository{db:db}
}

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

    // 3. Generate Dummy Seats (Misal 10 kursi saja dulu: A1 - A10)
    // Di dunia nyata, ini logic kompleks. Disini kita loop sederhana.
    querySeat := `INSERT INTO seats (event_id, seat_number, is_booked) VALUES ($1, $2, False)`
    
    for i := 1; i <= 10; i++ {
        seatNum := fmt.Sprintf("A%d", i) // A1, A2...
        _, err := tx.Exec(ctx, querySeat, event.ID, seatNum)
        if err != nil {
            return err // Jika gagal generate kursi, Event juga batal dibuat (Rollback otomatis)
        }
    }

    // 4. Commit Transaksi (Simpan Permanen)
    return tx.Commit(ctx)
}

func (r *eventRepository) GetAllEvents(ctx context.Context) ([]entity.Event, error){
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

	return events, nil 
}