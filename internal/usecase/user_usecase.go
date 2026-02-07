package usecase

import (
	"context"
	"errors"
	"time"

	"ticres/internal/entity"
	"ticres/internal/repository"
	"ticres/pkg/logger"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
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

func (uc *userUsecase) Register(ctx context.Context, user *entity.User) error {
	logger.Debug("registering new user", logger.String("email", user.Email))

	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("failed to hash password", logger.Err(err))
		return err
	}

	user.Password = string(hashedPassword)

	err = uc.userRepo.CreateUser(ctx, user)
	if err != nil {
		logger.Error("failed to create user",
			logger.String("email", user.Email),
			logger.Err(err),
		)
		return err
	}

	logger.Info("user registered successfully",
		logger.Int64("user_id", user.ID),
		logger.String("email", user.Email),
	)
	return nil
}

func (uc *userUsecase) Login(ctx context.Context, email, password string) (string, error) {
	logger.Debug("user login attempt", logger.String("email", email))

	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	user, err := uc.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		logger.Warn("login failed: user not found", logger.String("email", email))
		return "", entity.ErrInternalServer
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		logger.Warn("login failed: invalid password", logger.String("email", email))
		return "", errors.New("invalid email or password")
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Duration(uc.jwtExp) * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(uc.jwtSecret))
	if err != nil {
		logger.Error("failed to sign JWT token", logger.Err(err))
		return "", err
	}

	logger.Info("user logged in successfully",
		logger.Int64("user_id", user.ID),
		logger.String("email", email),
		logger.String("role", user.Role),
	)
	return signedToken, nil
}

func (uc *userUsecase) GetProfile(ctx context.Context, userID int) (*entity.User, error) {
	logger.Debug("fetching user profile", logger.Int("user_id", userID))

	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	user, err := uc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		logger.Warn("failed to get user profile", logger.Int("user_id", userID), logger.Err(err))
		return nil, err
	}

	logger.Debug("user profile fetched", logger.Int("user_id", userID))
	return user, nil
}