package http

import (
	"net/http"

	"ticres/internal/entity"
	"ticres/internal/usecase"
	"ticres/pkg/logger"

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
	logger.Debug("handler: register request received")

	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("handler: invalid register request", logger.Err(err))
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
			logger.Warn("handler: registration failed - email already exists", logger.String("email", req.Email))
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}
		logger.Error("handler: registration failed", logger.String("email", req.Email), logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal registrasi user: " + err.Error()})
		return
	}

	logger.Info("handler: user registered successfully",
		logger.Int64("user_id", user.ID),
		logger.String("email", user.Email),
	)
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
	logger.Debug("handler: login request received")

	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("handler: invalid login request", logger.Err(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.userUsecase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err.Error() == "invalid email or password" {
			logger.Warn("handler: login failed - invalid credentials", logger.String("email", req.Email))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		} else {
			logger.Error("handler: login failed", logger.String("email", req.Email), logger.Err(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		}
		return
	}

	logger.Info("handler: user logged in", logger.String("email", req.Email))
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func (h *UserHandler) Me(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		logger.Warn("handler: user not authenticated for /me endpoint")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid := int(userID.(float64))
	logger.Debug("handler: fetching user profile", logger.Int("user_id", uid))

	user, err := h.userUsecase.GetProfile(c.Request.Context(), uid)
	if err != nil {
		logger.Error("handler: failed to get user profile", logger.Int("user_id", uid), logger.Err(err))
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
		logger.Warn("handler: user not authenticated for /me/bookings endpoint")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid := int64(userID.(float64))
	logger.Debug("handler: fetching user bookings", logger.Int64("user_id", uid))

	bookings, err := h.bookingUsecase.GetBookingsByUserID(c.Request.Context(), uid)
	if err != nil {
		logger.Error("handler: failed to get user bookings", logger.Int64("user_id", uid), logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get bookings"})
		return
	}

	logger.Debug("handler: user bookings fetched", logger.Int64("user_id", uid), logger.Int("count", len(bookings)))
	c.JSON(http.StatusOK, gin.H{
		"data": bookings,
	})
}
