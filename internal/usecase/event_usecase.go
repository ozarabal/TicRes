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
}

type eventUsecase struct {
	eventRepo      repository.EventRepository
	contextTimeout time.Duration
}

func NewEventUsecase(repo repository.EventRepository, timeout time.Duration) EventUsecase {
	return &eventUsecase{eventRepo: repo, contextTimeout: timeout}
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