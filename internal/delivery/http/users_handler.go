package http

import (
	"net/http"

	"ticres/internal/entity"
	"ticres/internal/usecase"

	"github.com/gin-gonic/gin"
)

// 1. Struct Handler (Menyimpan dependency ke Usecase)
type UserHandler struct {
	userUsecase usecase.UserUsecase
}

// Constructor
func NewUserHandler(userUsecase usecase.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase}
}

// 2. Definisi Struktur Request (DTO - Data Transfer Object)
// Kita pisahkan struct ini dari Entity database agar validasinya spesifik untuk Register.
type registerRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// 3. Fungsi Handle Register
func (h *UserHandler) Register(c *gin.Context) {
	var req registerRequest

	// A. Bind JSON & Validasi
	// Jika JSON tidak sesuai struct (misal: email kosong), otomatis error.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// B. Mapping dari DTO ke Entity
	user := &entity.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	// C. Panggil Logic Bisnis (Usecase)
	// Kita gunakan Context dari Gin (c.Request.Context()) agar trace-nya nyambung
	if err := h.userUsecase.Register(c.Request.Context(), user); err != nil {

		if err == entity.ErrUserAlreadyExsist{
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}
		// Disini kita bisa cek error type, tapi untuk simpelnya kita return 500 dulu
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal registrasi user: " + err.Error()})
		return
	}

	// D. Sukses
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

// 1. Buat DTO untuk Login Request
type loginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

// 2. Tambahkan Fungsi Login
func (h *UserHandler) Login(c *gin.Context) {
    var req loginRequest

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    token, err := h.userUsecase.Login(c.Request.Context(), req.Email, req.Password)
    if err != nil {
        // Bedakan error validasi login vs error server
        if err.Error() == "invalid email or password" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
        }
        return
    }

    // Sukses return token
    c.JSON(http.StatusOK, gin.H{
        "token": token,
    })
}

func (h *UserHandler) Me(c *gin.Context) {
	// Ambil userID yang tadi disimpan oleh Middleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ini adalah halaman rahasia",
		"user_id": userID, // Bukti bahwa kita tahu siapa user yang login
	})
}