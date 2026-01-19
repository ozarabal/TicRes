package usecase

import (
	"context"
	"time"

	"ticres/internal/entity"
	"ticres/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// 1. Definisi Interface
type UserUsecase interface {
	Register(ctx context.Context, user *entity.User) error
}

// 2. Struct Implementasi
type userUsecase struct {
	userRepo       repository.UserRepository
	contextTimeout time.Duration
}

// Constructor
func NewUserUsecase(u repository.UserRepository, timeout time.Duration) UserUsecase {
	return &userUsecase{
		userRepo:       u,
		contextTimeout: timeout,
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