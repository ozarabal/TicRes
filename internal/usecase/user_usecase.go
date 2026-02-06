package usecase

import (
	"context"
	"time"

	"errors"

	"ticres/internal/entity"
	"ticres/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
)

type UserUsecase interface {
	Register(ctx context.Context, user *entity.User) error
	Login(ctx context.Context, email string, password string) (string, error)
	GetProfile(ctx context.Context, userID int) (*entity.User, error)
}

// 2. Struct Implementasi
type userUsecase struct {
	userRepo       repository.UserRepository
	contextTimeout time.Duration
	jwtSecret		string
	jwtExp			int	
}

// Constructor
func NewUserUsecase(u repository.UserRepository, timeout time.Duration, jwtSecret string, jwtExp int) UserUsecase {
	return &userUsecase{
		userRepo:       u,
		contextTimeout: timeout,
		jwtSecret: jwtSecret,
		jwtExp: jwtExp,
	}
}

// 3. Implementasi Logic
func (uc *userUsecase) Register(ctx context.Context, user *entity.User) error {
	// A. Setup Timeout
	// Agar jika database hang, request user tidak menunggu selamanya.
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	// B. Hashing Password (Business Logic)
	// Kita ubah "rahasia123" menjadi "$2a$10$N9qo8uLOickgx2ZMRZoM..."
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	// Ganti password asli dengan yang sudah di-hash
	user.Password = string(hashedPassword)

	// C. Panggil Repository (Data Layer)
	err = uc.userRepo.CreateUser(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (uc *userUsecase) Login(ctx context.Context, email, password string) (string, error) {
    ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
    defer cancel()

    // 1. Cari User by Email
    user, err := uc.userRepo.GetUserByEmail(ctx, email)
    if err != nil {
        // Best Practice: Jangan beri tahu email tidak ditemukan (security)
        // Tapi untuk debug boleh return err dulu. Idealnya return "Invalid email or password"
        return "", entity.ErrInternalServer
    }

    // 2. Verifikasi Password (Hash vs Plain)
    err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
    if err != nil {
        return "", errors.New("invalid email or password")
    }

    // 3. Generate JWT Token
    // Claims adalah data yang mau kita simpan di dalam token (Payload)
    claims := jwt.MapClaims{
        "user_id": user.ID,
        "email":   user.Email,
		"role" : user.Role,
        "exp":     time.Now().Add(time.Duration(uc.jwtExp) * time.Hour).Unix(), // Expired kapan
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    // Tanda tangani token dengan Secret Key
    signedToken, err := token.SignedString([]byte(uc.jwtSecret))
    if err != nil {
        return "", err
    }

    return signedToken, nil
}

func (uc *userUsecase) GetProfile(ctx context.Context, userID int) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	return uc.userRepo.GetUserByID(ctx, userID)
}