package repository

import (
	"context"
	"ticres/internal/entity" // sesuaikan nama module
	"errors"
	
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// 1. Definisi Interface (Kontrak Kerja)
type UserRepository interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
}

// 2. Implementasi Interface
type userRepository struct {
	db *pgxpool.Pool
}

// Constructor untuk membuat repository
func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

// --- Implementasi Method ---

func (r *userRepository) CreateUser(ctx context.Context, user *entity.User) error {
	// TANTANGAN: Tulis Query Insert disini.
	// Gunakan 'Returning id, created_at' agar struct user terupdate dengan ID baru dari DB.
	
	query := `
		INSERT INTO users (name, username, email, password, created_at) 
		VALUES ($1, $2, $3, $4, NOW()) 
		RETURNING user_id, created_at
	`

	// Eksekusi query menggunakan r.db.QueryRow(...)
	// Scan hasilnya ke user.ID dan user.CreatedAt
	
	err := r.db.QueryRow(ctx, query, user.Name, user.UserName, user.Email, user.Password).Scan(&user.ID, &user.CreatedAt)
	if err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return entity.ErrUserAlreadyExsist
			}
		}

		return err
	}

	return nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	// TANTANGAN: Implementasikan Query Select disini
	var user entity.User
	
	query := `SELECT user_id, name,username, email, password, created_at FROM users WHERE email = $1`

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, 
		&user.Name,
		&user.UserName, 
		&user.Email, 
		&user.Password, 
		&user.CreatedAt,
	)
	
	if err != nil {
		return nil, err
	}

	return &user, nil
}