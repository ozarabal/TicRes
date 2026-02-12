package database

import (
	"context"
	"fmt"
	"time"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresConnection(host, port, user, password, dbname, sslmode string) (*pgxpool.Pool, error) {
	// 1. Format Connection String (DSN)
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode,
	)

	// 2. Parse Config
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// 3. Setup Connection Pool Settings (Penting untuk performa!)
	config.MaxConns = 10                       // Maksimal 10 koneksi terbuka
	config.MinConns = 2                        // Minimal 2 koneksi standby
	config.MaxConnLifetime = 1 * time.Hour     // Refresh koneksi setiap jam
	config.MaxConnIdleTime = 30 * time.Minute  // Tutup koneksi jika nganggur 30 menit

	// 4. Create Pool
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	// 5. Test Ping (Pastikan benar-benar connect)
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}