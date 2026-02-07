package http

import (
	"net/http"
	"strconv"
	"time"

	"ticres/internal/entity"
	"ticres/internal/usecase"
	"ticres/pkg/logger"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	eventUsecase usecase.EventUsecase
}

func NewEventHandler(u usecase.EventUsecase) *EventHandler {
	return &EventHandler{eventUsecase: u}
}

type createEventRequest struct {
	Name     string `json:"name" binding:"required"`
	Location string `json:"location" binding:"required"`
	Date     string `json:"date" binding:"required"`
	Capacity int    `json:"capacity" binding:"required,min=1"`
}

func (h *EventHandler) Create(c *gin.Context) {
	logger.Debug("handler: create event request received")

	var req createEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("handler: invalid create event request", logger.Err(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parsedDate, err := time.Parse("2006-01-02 15:04", req.Date)
	if err != nil {
		logger.Warn("handler: invalid date format", logger.String("date", req.Date))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD HH:MM"})
		return
	}

	event := &entity.Event{
		Name:     req.Name,
		Location: req.Location,
		Date:     parsedDate,
		Capacity: req.Capacity,
	}

	if err := h.eventUsecase.CreateEvent(c.Request.Context(), event); err != nil {
		logger.Error("handler: failed to create event", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("handler: event created", logger.Int64("event_id", event.ID), logger.String("name", event.Name))
	c.JSON(http.StatusCreated, event)
}

func (h *EventHandler) List(c *gin.Context) {
	search := c.Query("search")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	logger.Debug("handler: listing events",
		logger.String("search", search),
		logger.Int("page", page),
		logger.Int("limit", limit),
	)

	events, total, err := h.eventUsecase.ListEventsWithSearch(c.Request.Context(), search, page, limit)
	if err != nil {
		logger.Error("handler: failed to list events", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hasMore := (page * limit) < total

	logger.Debug("handler: events listed", logger.Int("total", total), logger.Int("returned", len(events)))
	c.JSON(http.StatusOK, gin.H{
		"data": events,
		"meta": gin.H{
			"total":   total,
			"page":    page,
			"limit":   limit,
			"hasMore": hasMore,
		},
	})
}

func (h *EventHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	eventID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		logger.Warn("handler: invalid event ID", logger.String("id", idParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	logger.Debug("handler: getting event by ID", logger.Int64("event_id", eventID))

	eventWithSeats, err := h.eventUsecase.GetEventWithSeats(c.Request.Context(), eventID)
	if err != nil {
		logger.Warn("handler: event not found", logger.Int64("event_id", eventID), logger.Err(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": eventWithSeats})
}

type updateEventRequest struct {
	Name     string `json:"name" binding:"required"`
	Location string `json:"location" binding:"required"`
	Date     string `json:"date" binding:"required"`
	Capacity int    `json:"capacity" binding:"required,min=1"`
}

func (h *EventHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	eventID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		logger.Warn("handler: invalid event ID for update", logger.String("id", idParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	logger.Debug("handler: update event request", logger.Int64("event_id", eventID))

	existingEvent, err := h.eventUsecase.GetEventByID(c.Request.Context(), eventID)
	if err != nil {
		logger.Warn("handler: event not found for update", logger.Int64("event_id", eventID))
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	var req updateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("handler: invalid update event request", logger.Err(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parsedDate, err := time.Parse("2006-01-02 15:04", req.Date)
	if err != nil {
		logger.Warn("handler: invalid date format for update", logger.String("date", req.Date))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD HH:MM"})
		return
	}

	event := &entity.Event{
		ID:        eventID,
		Name:      req.Name,
		Location:  req.Location,
		Date:      parsedDate,
		Capacity:  req.Capacity,
		UpdatedAt: time.Now(),
	}

	if err := h.eventUsecase.EditEvent(c.Request.Context(), event, int64(existingEvent.Capacity)); err != nil {
		logger.Error("handler: failed to update event", logger.Int64("event_id", eventID), logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("handler: event updated", logger.Int64("event_id", eventID))
	c.JSON(http.StatusOK, gin.H{
		"message": "Event updated successfully",
		"data":    event,
	})
}

func (h *EventHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	eventID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		logger.Warn("handler: invalid event ID for delete", logger.String("id", idParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	logger.Info("handler: cancelling event", logger.Int64("event_id", eventID))

	err = h.eventUsecase.CancelEvent(c.Request.Context(), eventID)
	if err != nil {
		logger.Error("handler: failed to cancel event", logger.Int64("event_id", eventID), logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("handler: event cancelled", logger.Int64("event_id", eventID))
	c.JSON(http.StatusOK, gin.H{
		"message": "Event cancelled. Refund process started in background.",
	})
}
