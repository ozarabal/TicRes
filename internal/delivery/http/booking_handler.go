package http

import (
	"fmt"
	"net/http"
	"ticres/internal/usecase"

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
	fmt.Print("userID: ",userIDFloat)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	}

	userID := int64(userIDFloat.(float64))

	var req bookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.bookingUC.BookSeats(c.Request.Context(), userID, req.EventID, req.SeatIDs)
	if err != nil {
        // Cek jika errornya karena kursi penuh
        if err.Error() == "seat not available or already booked" {
             c.JSON(http.StatusConflict, gin.H{"error": "Salah satu kursi yang dipilih sudah tidak tersedia"})
             return
        }
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Booking successful"})

}