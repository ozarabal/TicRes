package entity

import "time"

type Booking struct {
	ID        int64     `json:"booking_id"`
	UserID    int64     `json:"user_id"`
	EventID   int64     `json:"event_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Seat struct {
	ID         int64  `json:"seat_id"`
	EventID    int64  `json:"event_id"`
	SeatNumber string `json:"seat_number"`
	IsBooked   bool   `json:"is_booked"`
	Version    int    `json:"-"`
}

// BookingWithDetails includes event and user info for API responses
type BookingWithDetails struct {
	ID        int64     `json:"booking_id"`
	UserID    int64     `json:"user_id"`
	UserName  string    `json:"user_name"`
	UserEmail string    `json:"user_email"`
	EventID   int64     `json:"event_id"`
	EventName string    `json:"event_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// EventWithSeats includes seats info for booking page
type EventWithSeats struct {
	Event Event  `json:"event"`
	Seats []Seat `json:"seats"`
}