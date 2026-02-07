package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"ticres/internal/config"
	"ticres/pkg/database"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	pool, err := database.NewPostgresConnection(
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name,
	)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// --- Seed Admin Account ---
	adminPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	var adminID int
	err = pool.QueryRow(ctx,
		`INSERT INTO users (name, username, email, password, role)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (email) DO UPDATE SET role = 'admin'
		 RETURNING user_id`,
		"Admin", "admin", "admin@ticres.com", string(adminPassword), "admin",
	).Scan(&adminID)
	if err != nil {
		log.Fatalf("failed to seed admin: %v", err)
	}
	fmt.Printf("Seeded admin account: id=%d, email=admin@ticres.com, password=admin123\n", adminID)

	// --- Seed Events ---
	events := []struct {
		Name     string
		Date     time.Time
		Location string
		Capacity int
	}{
		{"Konser Coldplay", time.Now().AddDate(0, 1, 0), "Jakarta", 5},
		{"Festival Jazz", time.Now().AddDate(0, 2, 0), "Bandung", 3},
		{"Stand Up Comedy Night", time.Now().AddDate(0, 0, 14), "Surabaya", 4},
	}

	for _, e := range events {
		tx, err := pool.Begin(ctx)
		if err != nil {
			log.Fatalf("failed to begin tx: %v", err)
		}

		var eventID int
		err = tx.QueryRow(ctx,
			`INSERT INTO events (name, date, location, capacity, status)
			 VALUES ($1, $2, $3, $4, 'available')
			 RETURNING event_id`,
			e.Name, e.Date, e.Location, e.Capacity,
		).Scan(&eventID)
		if err != nil {
			tx.Rollback(ctx)
			log.Fatalf("failed to seed event %q: %v", e.Name, err)
		}

		for i := 1; i <= e.Capacity; i++ {
			seatNumber := fmt.Sprintf("%d-%d", eventID, i)
			_, err = tx.Exec(ctx,
				`INSERT INTO seats (event_id, seat_number, category, is_booked)
				 VALUES ($1, $2, 'regular', false)`,
				eventID, seatNumber,
			)
			if err != nil {
				tx.Rollback(ctx)
				log.Fatalf("failed to seed seat %s: %v", seatNumber, err)
			}
		}

		if err := tx.Commit(ctx); err != nil {
			log.Fatalf("failed to commit event %q: %v", e.Name, err)
		}

		fmt.Printf("Seeded event: id=%d, name=%q, capacity=%d, seats=%d\n", eventID, e.Name, e.Capacity, e.Capacity)
	}

	fmt.Println("Seeding completed successfully!")
}
