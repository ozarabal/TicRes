package usecase

import(
	"time"
	"context"
	"ticres/internal/repository"
	// "ticres/internal/worker"
)

type BookingUsecase interface {
	BookSeats(ctx context.Context, userID, eventID int64, seatIDs []int64, userEmail string) error
}

type NotificationService interface {
	SendNotification(bookingID int64, email, message string)
}

type bookingUsecase struct {
	bookingRepo    repository.BookingRepository
	contextTimeout time.Duration
	// notifWorker    *worker.NotificationWorker
	notifWorker    NotificationService

}

func NewBookingUsecase(repo repository.BookingRepository, timeout time.Duration, notifWorker NotificationService) BookingUsecase {
	return &bookingUsecase{bookingRepo: repo, contextTimeout: timeout, notifWorker: notifWorker}
}

func (uc *bookingUsecase) BookSeats(ctx context.Context, userID, eventID int64, seatIDs []int64, useremail string) error {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	bookingID, err := uc.bookingRepo.CreateBooking(ctx, userID, eventID, seatIDs)
	if err != nil{
		return err
	}

	uc.notifWorker.SendNotification(bookingID,useremail, "Booking Berhasil!")

	return nil
}