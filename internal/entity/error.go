package entity

import "errors"

var (
	ErrUserAlreadyExsist   = errors.New("user with this email already exisist")
	ErrInternalServer      = errors.New("internal server error")
	ErrNotFound            = errors.New("data not found")
	ErrBookingNotPending   = errors.New("booking is not in PENDING state")
	ErrBookingExpired      = errors.New("booking has expired")
	ErrPaymentAlreadyMade  = errors.New("payment has already been completed")
	ErrInvalidPaymentMethod = errors.New("invalid payment method")
	ErrUnauthorized        = errors.New("unauthorized access")
)