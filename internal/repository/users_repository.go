package repository

import (
	"context"
	"errors"

	"ticres/internal/entity"
	"ticres/pkg/logger"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserByID(ctx context.Context, id int) (*entity.User, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (name, username, email, password, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING user_id, created_at
	`

	logger.Debug("creating user",
		logger.String("email", user.Email),
		logger.String("name", user.Name),
	)

	err := r.db.QueryRow(ctx, query, user.Name, user.UserName, user.Email, user.Password).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				logger.Warn("user creation failed: duplicate email",
					logger.String("email", user.Email),
					logger.String("pg_code", pgErr.Code),
				)
				return entity.ErrUserAlreadyExsist
			}
		}

		logger.Error("user creation failed",
			logger.String("email", user.Email),
			logger.Err(err),
		)
		return err
	}

	logger.Info("user created successfully",
		logger.Int64("user_id", user.ID),
		logger.String("email", user.Email),
	)
	return nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User

	query := `SELECT user_id, name, username, email, password, role, created_at FROM users WHERE email = $1`

	logger.Debug("fetching user by email", logger.String("email", email))

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.UserName,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
	)

	if err != nil {
		logger.Warn("user not found by email",
			logger.String("email", email),
			logger.Err(err),
		)
		return nil, err
	}

	logger.Debug("user found", logger.Int64("user_id", user.ID))
	return &user, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, ID int) (*entity.User, error) {
	query := `SELECT user_id, name, username, email, password, role, created_at FROM users WHERE user_id = $1`

	var user entity.User

	logger.Debug("fetching user by ID", logger.Int("user_id", ID))

	err := r.db.QueryRow(ctx, query, ID).Scan(
		&user.ID,
		&user.Name,
		&user.UserName,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
	)

	if err != nil {
		logger.Warn("user not found by ID",
			logger.Int("user_id", ID),
			logger.Err(err),
		)
		return nil, err
	}

	logger.Debug("user found", logger.Int64("user_id", user.ID))
	return &user, nil
}
