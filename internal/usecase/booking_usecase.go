package usecase

import(
	"time"
	"context"
	"ticres/internal/repository"
)

type BookingUsecase interface {
	BookSeats(ctx context.Context, userID, eventID int64, seatIDs []int64) error
}

type bookingUsecase struct {
	bookingRepo    repository.BookingRepository
	contextTimeout time.Duration
}

func NewBookingUsecase(repo repository.BookingRepository, timeout time.Duration) BookingUsecase {
	return &bookingUsecase{bookingRepo: repo, contextTimeout: timeout}
}

func (uc *bookingUsecase) BookSeats(ctx context.Context, userID, eventID int64, seatIDs []int64) error {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	return uc.bookingRepo.CreateBooking(ctx, userID, eventID, seatIDs)
}