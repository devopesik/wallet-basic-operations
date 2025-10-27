package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/devopesik/wallet-basic-operations/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type walletRepository struct {
	pool *pgxpool.Pool
}

func NewWalletRepository(pool *pgxpool.Pool) repository.WalletRepository {
	return &walletRepository{pool: pool}
}

func (r *walletRepository) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	var balance int64
	err := r.pool.QueryRow(ctx, "SELECT balance FROM wallets WHERE id = $1", walletID).Scan(&balance)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("кошелёк не найден")
		}
		return 0, fmt.Errorf("ошибка базы данных при получении баланса: %w", err)
	}
	return balance, nil
}

func (r *walletRepository) UpdateBalance(ctx context.Context, walletID uuid.UUID, amount int64, isDeposit bool) error {
	if amount <= 0 {
		return fmt.Errorf("сумма должна быть положительной")
	}

	var result pgconn.CommandTag
	var err error

	if isDeposit {
		query := "UPDATE wallets SET balance = balance + $1 WHERE id = $2"
		result, err = r.pool.Exec(ctx, query, amount, walletID)
	} else {
		query := "UPDATE wallets SET balance = balance - $1 WHERE id = $2 AND balance >= $1"
		result, err = r.pool.Exec(ctx, query, amount, walletID)
	}

	if err != nil {
		return fmt.Errorf("ошибка базы данных при обновлении баланса: %w", err)
	}

	if result.RowsAffected() == 0 {
		if isDeposit {
			return fmt.Errorf("кошелёк не найден")
		} else {
			return fmt.Errorf("недостаточно средств")
		}
	}

	return nil
}

func (r *walletRepository) CreateWallet(ctx context.Context, walletID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "INSERT INTO wallets (id, balance) VALUES ($1, 0)", walletID)
	var pgErr *pgconn.PgError
	if err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return fmt.Errorf("кошелёк уже существует")
		}
		return fmt.Errorf("ошибка базы данных при создании кошелька: %w", err)
	}
	return nil
}
