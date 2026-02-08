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
	Name        string  `json:"name" binding:"required"`
	Location    string  `json:"location" binding:"required"`
	Date        string  `json:"date" binding:"required"`
	Capacity    int     `json:"capacity" binding:"required,min=1"`
	TicketPrice float64 `json:"ticket_price" binding:"required,min=0"`
}

// Create godoc
// @Summary      Create a new event
// @Description  Create a new event with details and ticket price. Authenticated user required.
// @Tags         events
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body createEventRequest true "Event creation details"
// @Success      201 {object} entity.Event "Event created successfully"
// @Failure      400 {object} map[string]string "Invalid request body or date format"
// @Failure      401 {object} map[string]string "User not authenticated"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /events [post]
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

	if err := h.eventUsecase.CreateEvent(c.Request.Context(), event, req.TicketPrice); err != nil {
		logger.Error("handler: failed to create event", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("handler: event created", logger.Int64("event_id", event.ID), logger.String("name", event.Name))
	c.JSON(http.StatusCreated, event)
}

// List godoc
// @Summary      List events
// @Description  Retrieve a paginated list of events with optional search filter
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        search query string false "Search by event name or location"
// @Param        page query int false "Page number" default(1) minimum(1)
// @Param        limit query int false "Items per page (max 100)" default(10) minimum(1) maximum(100)
// @Success      200 {object} map[string]interface{} "List of events with pagination metadata"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /events [get]
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

// GetByID godoc
// @Summary      Get event by ID
// @Description  Retrieve detailed information about a specific event including available seats
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        id path int true "Event ID" example(1)
// @Success      200 {object} map[string]interface{} "Event details with seats information"
// @Failure      400 {object} map[string]string "Invalid event ID"
// @Failure      404 {object} map[string]string "Event not found"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /events/{id} [get]
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

// Update godoc
// @Summary      Update an event
// @Description  Update event details. Admin access required. Capacity changes will create/delete seats accordingly.
// @Tags         events
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Event ID" example(1)
// @Param        request body updateEventRequest true "Event update details"
// @Success      200 {object} map[string]interface{} "Event updated successfully"
// @Failure      400 {object} map[string]string "Invalid request or date format"
// @Failure      401 {object} map[string]string "User not authenticated"
// @Failure      403 {object} map[string]string "Access forbidden - admin only"
// @Failure      404 {object} map[string]string "Event not found"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /events/{id} [put]
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

// Delete godoc
// @Summary      Cancel an event
// @Description  Cancel an event and start automatic refund process for all bookings. Admin access required.
// @Tags         events
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Event ID" example(1)
// @Success      200 {object} map[string]string "Event cancelled successfully, refund process started"
// @Failure      400 {object} map[string]string "Invalid event ID"
// @Failure      401 {object} map[string]string "User not authenticated"
// @Failure      403 {object} map[string]string "Access forbidden - admin only"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /events/{id} [delete]
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
