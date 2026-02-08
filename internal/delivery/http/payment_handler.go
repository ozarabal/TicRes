package http

import (
	"errors"
	"net/http"
	"strconv"

	"ticres/internal/entity"
	"ticres/internal/usecase"
	"ticres/pkg/logger"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentUC usecase.PaymentUsecase
}

func NewPaymentHandler(uc usecase.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{paymentUC: uc}
}

type payRequest struct {
	BookingID     int64  `json:"booking_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required,oneof=credit_card bank_transfer e_wallet"`
}

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	userIDFloat, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := int64(userIDFloat.(float64))

	var req payRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("handler: invalid payment request", logger.Err(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Info("handler: processing payment",
		logger.Int64("user_id", userID),
		logger.Int64("booking_id", req.BookingID),
		logger.String("payment_method", req.PaymentMethod),
	)

	txn, err := h.paymentUC.ProcessPayment(c.Request.Context(), req.BookingID, userID, req.PaymentMethod)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		case errors.Is(err, entity.ErrUnauthorized):
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this booking"})
		case errors.Is(err, entity.ErrBookingExpired):
			c.JSON(http.StatusGone, gin.H{"error": "Booking has expired. Please create a new booking."})
		case errors.Is(err, entity.ErrPaymentAlreadyMade):
			c.JSON(http.StatusConflict, gin.H{"error": "Payment has already been completed for this booking"})
		case errors.Is(err, entity.ErrBookingNotPending):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Booking is not in a payable state"})
		case errors.Is(err, entity.ErrInvalidPaymentMethod):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method. Use: credit_card, bank_transfer, or e_wallet"})
		default:
			logger.Error("handler: payment processing failed", logger.Err(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment processing failed"})
		}
		return
	}

	logger.Info("handler: payment successful",
		logger.Int64("booking_id", req.BookingID),
		logger.String("external_id", txn.ExternalID),
	)
	c.JSON(http.StatusOK, gin.H{
		"message": "Payment successful",
		"data":    txn,
	})
}

func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
	userIDFloat, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := int64(userIDFloat.(float64))

	bookingIDStr := c.Param("booking_id")
	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	result, err := h.paymentUC.GetPaymentStatus(c.Request.Context(), bookingID, userID)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		case errors.Is(err, entity.ErrUnauthorized):
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this booking"})
		default:
			logger.Error("handler: failed to get payment status", logger.Err(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get payment status"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}
