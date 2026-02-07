package http

import (
	"net/http"
	"strconv"

	"ticres/internal/usecase"
	"ticres/pkg/logger"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	bookingUsecase usecase.BookingUsecase
}

func NewAdminHandler(bookingUsecase usecase.BookingUsecase) *AdminHandler {
	return &AdminHandler{bookingUsecase: bookingUsecase}
}

func (h *AdminHandler) GetAllBookings(c *gin.Context) {
	status := c.Query("status")
	sortBy := c.DefaultQuery("sort", "created_at")
	sortOrder := c.DefaultQuery("order", "desc")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	logger.Debug("handler: admin fetching all bookings",
		logger.String("status", status),
		logger.Int("page", page),
		logger.Int("limit", limit),
	)

	bookings, total, err := h.bookingUsecase.GetAllBookings(c.Request.Context(), status, sortBy, sortOrder, page, limit)
	if err != nil {
		logger.Error("handler: admin failed to get all bookings", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hasMore := (page * limit) < total

	logger.Debug("handler: admin bookings fetched", logger.Int("total", total), logger.Int("returned", len(bookings)))
	c.JSON(http.StatusOK, gin.H{
		"data": bookings,
		"meta": gin.H{
			"total":   total,
			"page":    page,
			"limit":   limit,
			"hasMore": hasMore,
		},
	})
}

func (h *AdminHandler) GetEventBookings(c *gin.Context) {
	idParam := c.Param("id")
	eventID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		logger.Warn("handler: admin invalid event ID", logger.String("id", idParam))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	status := c.Query("status")
	sortBy := c.DefaultQuery("sort", "created_at")
	sortOrder := c.DefaultQuery("order", "desc")

	logger.Debug("handler: admin fetching event bookings",
		logger.Int64("event_id", eventID),
		logger.String("status", status),
	)

	bookings, err := h.bookingUsecase.GetBookingsByEventID(c.Request.Context(), eventID, status, sortBy, sortOrder)
	if err != nil {
		logger.Error("handler: admin failed to get event bookings",
			logger.Int64("event_id", eventID),
			logger.Err(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Debug("handler: admin event bookings fetched",
		logger.Int64("event_id", eventID),
		logger.Int("count", len(bookings)),
	)
	c.JSON(http.StatusOK, gin.H{
		"data": bookings,
	})
}
