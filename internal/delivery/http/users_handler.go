package http

import (
	"net/http"

	"ticres/internal/entity"
	"ticres/internal/usecase"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUsecase    usecase.UserUsecase
	bookingUsecase usecase.BookingUsecase
}

func NewUserHandler(userUsecase usecase.UserUsecase, bookingUsecase usecase.BookingUsecase) *UserHandler {
	return &UserHandler{
		userUsecase:    userUsecase,
		bookingUsecase: bookingUsecase,
	}
}

type registerRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var req registerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	if err := h.userUsecase.Register(c.Request.Context(), user); err != nil {
		if err == entity.ErrUserAlreadyExsist {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal registrasi user: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"data": gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"created_at": user.CreatedAt,
		},
	})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *UserHandler) Login(c *gin.Context) {
	var req loginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.userUsecase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err.Error() == "invalid email or password" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func (h *UserHandler) Me(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := h.userUsecase.GetProfile(c.Request.Context(), int(userID.(float64)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

func (h *UserHandler) GetMyBookings(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	bookings, err := h.bookingUsecase.GetBookingsByUserID(c.Request.Context(), int64(userID.(float64)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get bookings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": bookings,
	})
}
