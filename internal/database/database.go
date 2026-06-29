package database

import (
	"context"
	"fmt"
	"iamsainadh/medassist_ai/internal/config"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresConnection(cfg *config.Config) (*pgxpool.Pool, error) {
	currCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	defer cancel()

	pool, err := pgxpool.New(currCtx, cfg.DBUri)

	if err != nil {
		return nil, fmt.Errorf("unable to create db pool: %w", err)

	}

	if err := pool.Ping(currCtx); err != nil {
		defer pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	fmt.Println("Connected to PostgreSQL!")

	return pool, nil
}
