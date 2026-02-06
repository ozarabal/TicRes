package usecase

import (
	"context"
	"time"

	"ticres/internal/entity"
	"ticres/internal/repository"
)

type EventUsecase interface {
	CreateEvent(ctx context.Context, event *entity.Event) error
	ListEvents(ctx context.Context) ([]entity.Event, error)
	ListEventsWithSearch(ctx context.Context, search string, page, limit int) ([]entity.Event, int, error)
	GetEventByID(ctx context.Context, eventID int64) (*entity.Event, error)
	GetEventWithSeats(ctx context.Context, eventID int64) (*entity.EventWithSeats, error)
	EditEvent(ctx context.Context, event *entity.Event, prev int64) error
	CancelEvent(ctx context.Context, eventID int64) error
}

type eventUsecase struct {
	eventRepo      repository.EventRepository
	contextTimeout time.Duration
	worker			NotificationService
}

func NewEventUsecase(repo repository.EventRepository, timeout time.Duration, worker NotificationService) EventUsecase {
	return &eventUsecase{eventRepo: repo, contextTimeout: timeout, worker: worker}
}

func (uc *eventUsecase) CreateEvent(ctx context.Context, event *entity.Event) error {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	return uc.eventRepo.CreateEvent(ctx, event)
}

func (uc *eventUsecase) ListEvents(ctx context.Context) ([]entity.Event, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	return uc.eventRepo.GetAllEvents(ctx)
}

func (uc *eventUsecase) ListEventsWithSearch(ctx context.Context, search string, page, limit int) ([]entity.Event, int, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	return uc.eventRepo.GetEventsWithSearch(ctx, search, page, limit)
}

func (uc *eventUsecase) GetEventByID(ctx context.Context, eventID int64) (*entity.Event, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	return uc.eventRepo.GetEventByID(ctx, eventID)
}

func (uc *eventUsecase) GetEventWithSeats(ctx context.Context, eventID int64) (*entity.EventWithSeats, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	return uc.eventRepo.GetEventWithSeats(ctx, eventID)
}

func (uc *eventUsecase) EditEvent(ctx context.Context, event *entity.Event, prev int64) error {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	return uc.eventRepo.UpdateEvent(ctx, event, prev)
}

func (uc *eventUsecase) CancelEvent(ctx context.Context, eventID int64) error {
    ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
    defer cancel()

    // 1. Cek apakah event ada (Opsional, tapi baik untuk validasi)
    // ...

    // 2. Update Status Event -> CANCELLED (Synchronous)
    // Agar user tidak bisa booking lagi detik ini juga.
    err := uc.eventRepo.UpdateEventStatus(ctx, eventID, "CANCELLED")
    if err != nil {
        return err
    }

    // 3. Trigger Refund Process di Background (Asynchronous)
    // Gunakan method baru di worker
    uc.worker.EnqueueCancellation(eventID)

    return nil
}