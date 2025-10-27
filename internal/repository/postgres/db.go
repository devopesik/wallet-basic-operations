package postgres

import (
	"context"
	"fmt"
	"log"

	"github.com/devopesik/wallet-basic-operations/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(cfg *config.Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать пул pgx: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("не удалось подключиться к БД через pgx: %w", err)
	}

	log.Println("Пул pgx создан")
	return pool, nil
}
