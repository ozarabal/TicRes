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
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode,
	)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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
		Price    float64
		Category string
	}{
		// Concerts
		{"Konser Coldplay Jakarta 2026", time.Now().AddDate(0, 1, 0), "Jakarta", 50, 1500000, "vip"},
		{"Konser Tulus - Manusia", time.Now().AddDate(0, 1, 15), "Bandung", 30, 500000, "regular"},
		{"Raisa Live in Concert", time.Now().AddDate(0, 2, 0), "Surabaya", 40, 750000, "regular"},
		{"Dewa 19 Reunion Tour", time.Now().AddDate(0, 2, 10), "Yogyakarta", 35, 600000, "regular"},
		{"Noah Band Anniversary", time.Now().AddDate(0, 3, 0), "Semarang", 25, 400000, "regular"},

		// Festivals
		{"Jakarta International Jazz Festival", time.Now().AddDate(0, 1, 20), "Jakarta", 100, 2000000, "vip"},
		{"Bali Spirit Festival", time.Now().AddDate(0, 2, 5), "Bali", 60, 1000000, "regular"},
		{"We The Fest 2026", time.Now().AddDate(0, 3, 10), "Jakarta", 80, 1800000, "vip"},
		{"Soundrenaline Bali", time.Now().AddDate(0, 4, 0), "Bali", 70, 900000, "regular"},
		{"Synchronize Fest", time.Now().AddDate(0, 2, 20), "Jakarta", 90, 750000, "regular"},

		// Comedy & Theater
		{"Stand Up Comedy: Raditya Dika", time.Now().AddDate(0, 0, 14), "Jakarta", 20, 350000, "regular"},
		{"Teater Koma: Semar Mesem", time.Now().AddDate(0, 1, 5), "Jakarta", 15, 250000, "regular"},
		{"Comedy Night Surabaya", time.Now().AddDate(0, 0, 21), "Surabaya", 18, 200000, "regular"},
		{"Improv Comedy Show", time.Now().AddDate(0, 1, 10), "Bandung", 12, 150000, "regular"},

		// Sports
		{"Indonesia Open Badminton 2026", time.Now().AddDate(0, 3, 5), "Jakarta", 45, 500000, "regular"},
		{"Persija vs Persib - Liga 1", time.Now().AddDate(0, 0, 7), "Jakarta", 60, 200000, "regular"},
		{"Jakarta Marathon 2026", time.Now().AddDate(0, 4, 15), "Jakarta", 200, 350000, "regular"},

		// Conferences & Workshops
		{"GoTo Tech Conference", time.Now().AddDate(0, 2, 0), "Jakarta", 30, 1500000, "vip"},
		{"Startup Summit Indonesia", time.Now().AddDate(0, 2, 15), "Bali", 25, 1000000, "regular"},
		{"DevFest Surabaya 2026", time.Now().AddDate(0, 1, 25), "Surabaya", 20, 0, "regular"},
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
				`INSERT INTO seats (event_id, seat_number, category, is_booked, price)
				 VALUES ($1, $2, $3, false, $4)`,
				eventID, seatNumber, e.Category, e.Price,
			)
			if err != nil {
				tx.Rollback(ctx)
				log.Fatalf("failed to seed seat %s: %v", seatNumber, err)
			}
		}

		if err := tx.Commit(ctx); err != nil {
			log.Fatalf("failed to commit event %q: %v", e.Name, err)
		}

		fmt.Printf("Seeded event: id=%d, name=%q, location=%s, capacity=%d\n", eventID, e.Name, e.Location, e.Capacity)
	}

	fmt.Println("Seeding completed successfully!")
}
