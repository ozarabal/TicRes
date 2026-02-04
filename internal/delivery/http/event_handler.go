package http

import (
	"net/http"
	"strconv"
	"time"

	"ticres/internal/entity"
	"ticres/internal/usecase"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	eventUsecase usecase.EventUsecase
}

func NewEventHandler(u usecase.EventUsecase) *EventHandler {
	return &EventHandler{eventUsecase: u}
}

// Struct request khusus untuk menangani format tanggal string
type createEventRequest struct {
	Name     string `json:"name" binding:"required"`
	Location string `json:"location" binding:"required"`
	Date     string `json:"date" binding:"required"` // Format: YYYY-MM-DD HH:MM
	Capacity int    `json:"capacity" binding:"required,min=1"`
}

func (h *EventHandler) Create(c *gin.Context) {
	var req createEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parsing string tanggal ke time.Time
	// Layout referensi Go itu unik: "2006-01-02 15:04" (Jan 2 15:04:05 2006 MST)
	parsedDate, err := time.Parse("2006-01-02 15:04", req.Date)
	if err != nil {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, event)
}

func (h *EventHandler) List(c *gin.Context) {
	events, err := h.eventUsecase.ListEvents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": events})
}

func (h *EventHandler) Delete(c *gin.Context) {
	// Ambil ID dari URL param /events/:id
	idParam := c.Param("id")
	eventID, _ := strconv.ParseInt(idParam, 10, 64)

	err := h.eventUsecase.CancelEvent(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Event cancelled. Refund process started in background.",
	})
}
