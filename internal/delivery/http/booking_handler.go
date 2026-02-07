package http

import (
	"net/http"

	"ticres/internal/usecase"
	"ticres/pkg/logger"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	bookingUC usecase.BookingUsecase
}

func NewBookingHandler(uc usecase.BookingUsecase) *BookingHandler {
	return &BookingHandler{bookingUC: uc}
}

type bookRequest struct {
	EventID int64   `json:"event_id" binding:"required"`
	SeatIDs []int64 `json:"seat_ids" binding:"required,min=1"`
}

func (h *BookingHandler) Create(c *gin.Context) {
	userIDFloat, exists := c.Get("userID")
	if !exists {
		logger.Warn("handler: unauthorized booking attempt")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID := int64(userIDFloat.(float64))
	userEmail := "customer@example.com"

	logger.Debug("handler: booking request received", logger.Int64("user_id", userID))

	var req bookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("handler: invalid booking request", logger.Err(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Debug("handler: booking seats",
		logger.Int64("user_id", userID),
		logger.Int64("event_id", req.EventID),
		logger.Int("seat_count", len(req.SeatIDs)),
	)

	err := h.bookingUC.BookSeats(c.Request.Context(), userID, req.EventID, req.SeatIDs, userEmail)
	if err != nil {
		if err.Error() == "seat not available or already booked" {
			logger.Warn("handler: booking failed - seat not available",
				logger.Int64("user_id", userID),
				logger.Int64("event_id", req.EventID),
			)
			c.JSON(http.StatusConflict, gin.H{"error": "Salah satu kursi yang dipilih sudah tidak tersedia"})
			return
		}
		logger.Error("handler: booking failed",
			logger.Int64("user_id", userID),
			logger.Int64("event_id", req.EventID),
			logger.Err(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Info("handler: booking created",
		logger.Int64("user_id", userID),
		logger.Int64("event_id", req.EventID),
		logger.Int("seat_count", len(req.SeatIDs)),
	)
	c.JSON(http.StatusCreated, gin.H{"message": "Booking successful, check your email"})
}