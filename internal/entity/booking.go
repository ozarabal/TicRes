package entity

import "time"

type Booking struct {
	ID        int64     `json:"booking_id"`
	UserID    int64     `json:"user_id"`
	EventID   int64     `json:"event_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	// Kita bisa tambahkan detail seats nanti jika perlu
}

type Seat struct {
	ID         int64  `json:"seat_id"`
	EventID    int64  `json:"event_id"`
	SeatNumber string `json:"seat_number"`
	IsBooked   bool   `json:"is_booked"`
	Version    int    `json:"-"` // Internal use only
}