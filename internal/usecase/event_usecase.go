package usecase

import (
	"context"
	"time"

	"ticres/internal/entity"
	"ticres/internal/repository"
	"ticres/pkg/logger"
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
	logger.Debug("usecase: creating event", logger.String("name", event.Name))

	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	err := uc.eventRepo.CreateEvent(ctx, event)
	if err != nil {
		logger.Error("usecase: failed to create event", logger.Err(err))
		return err
	}

	logger.Info("usecase: event created", logger.Int64("event_id", event.ID))
	return nil
}

func (uc *eventUsecase) ListEvents(ctx context.Context) ([]entity.Event, error) {
	logger.Debug("usecase: listing all events")

	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	events, err := uc.eventRepo.GetAllEvents(ctx)
	if err != nil {
		logger.Error("usecase: failed to list events", logger.Err(err))
		return nil, err
	}

	logger.Debug("usecase: events listed", logger.Int("count", len(events)))
	return events, nil
}

func (uc *eventUsecase) ListEventsWithSearch(ctx context.Context, search string, page, limit int) ([]entity.Event, int, error) {
	logger.Debug("usecase: listing events with search",
		logger.String("search", search),
		logger.Int("page", page),
		logger.Int("limit", limit),
	)

	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	events, total, err := uc.eventRepo.GetEventsWithSearch(ctx, search, page, limit)
	if err != nil {
		logger.Error("usecase: failed to search events", logger.Err(err))
		return nil, 0, err
	}

	logger.Debug("usecase: events search completed", logger.Int("total", total))
	return events, total, nil
}

func (uc *eventUsecase) GetEventByID(ctx context.Context, eventID int64) (*entity.Event, error) {
	logger.Debug("usecase: getting event by ID", logger.Int64("event_id", eventID))

	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	event, err := uc.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		logger.Warn("usecase: event not found", logger.Int64("event_id", eventID), logger.Err(err))
		return nil, err
	}

	return event, nil
}

func (uc *eventUsecase) GetEventWithSeats(ctx context.Context, eventID int64) (*entity.EventWithSeats, error) {
	logger.Debug("usecase: getting event with seats", logger.Int64("event_id", eventID))

	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	eventWithSeats, err := uc.eventRepo.GetEventWithSeats(ctx, eventID)
	if err != nil {
		logger.Warn("usecase: event with seats not found", logger.Int64("event_id", eventID), logger.Err(err))
		return nil, err
	}

	return eventWithSeats, nil
}

func (uc *eventUsecase) EditEvent(ctx context.Context, event *entity.Event, prev int64) error {
	logger.Debug("usecase: editing event",
		logger.Int64("event_id", event.ID),
		logger.Int64("prev_capacity", prev),
	)

	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	err := uc.eventRepo.UpdateEvent(ctx, event, prev)
	if err != nil {
		logger.Error("usecase: failed to edit event", logger.Int64("event_id", event.ID), logger.Err(err))
		return err
	}

	logger.Info("usecase: event edited", logger.Int64("event_id", event.ID))
	return nil
}

func (uc *eventUsecase) CancelEvent(ctx context.Context, eventID int64) error {
	logger.Info("usecase: cancelling event", logger.Int64("event_id", eventID))

	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	err := uc.eventRepo.UpdateEventStatus(ctx, eventID, "CANCELLED")
	if err != nil {
		logger.Error("usecase: failed to cancel event", logger.Int64("event_id", eventID), logger.Err(err))
		return err
	}

	uc.worker.EnqueueCancellation(eventID)
	logger.Info("usecase: event cancelled, refund process enqueued", logger.Int64("event_id", eventID))

	return nil
}