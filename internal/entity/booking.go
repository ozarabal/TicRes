package entity

import "time"

type Booking struct {
	ID          int64      `json:"booking_id"`
	UserID      int64      `json:"user_id"`
	EventID     int64      `json:"event_id"`
	Status      string     `json:"status"`
	TotalAmount float64    `json:"total_amount"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Seat struct {
	ID         int64   `json:"seat_id"`
	EventID    int64   `json:"event_id"`
	SeatNumber string  `json:"seat_number"`
	Category   string  `json:"category"`
	Price      float64 `json:"price"`
	IsBooked   bool    `json:"is_booked"`
	Version    int     `json:"-"`
}

type Transaction struct {
	ID              int64     `json:"payment_id"`
	Amount          float64   `json:"amount"`
	PaymentMethod   string    `json:"payment_method"`
	BookingID       int64     `json:"booking_id"`
	TransactionDate time.Time `json:"transaction_date"`
	ExternalID      string    `json:"external_id"`
	Status          string    `json:"status"`
}

type Refund struct {
	ID         int64     `json:"refund_id"`
	BookingID  int64     `json:"booking_id"`
	Amount     float64   `json:"amount"`
	RefundDate time.Time `json:"refund_date"`
	Reason     string    `json:"reason"`
	Status     string    `json:"status"`
}

// BookingWithPayment is the response for booking + payment info
type BookingWithPayment struct {
	BookingID   int64        `json:"booking_id"`
	EventID     int64        `json:"event_id"`
	Status      string       `json:"status"`
	TotalAmount float64      `json:"total_amount"`
	ExpiresAt   *time.Time   `json:"expires_at,omitempty"`
	Transaction *Transaction `json:"transaction,omitempty"`
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
