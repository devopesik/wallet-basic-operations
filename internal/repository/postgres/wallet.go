package postgres

import (
	"context"
	stderrors "errors"

	apperrors "github.com/devopesik/wallet-basic-operations/internal/errors"
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

func (r *walletRepository) GetWallet(ctx context.Context, walletID uuid.UUID) (*repository.Wallet, error) {
	var wallet repository.Wallet
	err := r.pool.QueryRow(ctx, "SELECT id, balance FROM wallets WHERE id = $1", walletID).Scan(&wallet.ID, &wallet.Balance)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.ErrWalletNotFound
		}
		return nil, apperrors.NewDatabaseError("получении кошелька", err)
	}
	return &wallet, nil
}

func (r *walletRepository) Deposit(ctx context.Context, walletID uuid.UUID, amount int64) error {
	// Начинаем транзакцию для атомарности операции
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return apperrors.NewDatabaseError("создание транзакции для пополнения", err)
	}
	defer tx.Rollback(ctx)

	// Обновляем баланс
	// UPDATE сам блокирует строку, поэтому SELECT FOR UPDATE не обязателен для Deposit
	query := "UPDATE wallets SET balance = balance + $1 WHERE id = $2"
	result, err := tx.Exec(ctx, query, amount, walletID)
	if err != nil {
		return apperrors.NewDatabaseError("пополнении баланса", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.ErrWalletNotFound
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return apperrors.NewDatabaseError("фиксация транзакции пополнения", err)
	}

	return nil
}

func (r *walletRepository) Withdraw(ctx context.Context, walletID uuid.UUID, amount int64) error {
	// Начинаем транзакцию для предотвращения race conditions
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return apperrors.NewDatabaseError("создание транзакции для списания", err)
	}
	defer tx.Rollback(ctx)

	var balance int64
	err = tx.QueryRow(ctx, "SELECT balance FROM wallets WHERE id = $1 FOR UPDATE", walletID).Scan(&balance)
	if err != nil {
		if err == pgx.ErrNoRows {
			return apperrors.ErrWalletNotFound
		}
		return apperrors.NewDatabaseError("получение баланса для списания", err)
	}

	// Проверяем достаточность средств
	if balance < amount {
		return apperrors.ErrInsufficientFunds
	}

	// Обновляем баланс
	query := "UPDATE wallets SET balance = balance - $1 WHERE id = $2"
	result, err := tx.Exec(ctx, query, amount, walletID)
	if err != nil {
		return apperrors.NewDatabaseError("списание баланса", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.ErrWalletNotFound
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return apperrors.NewDatabaseError("фиксация транзакции списания", err)
	}

	return nil
}

func (r *walletRepository) CreateWallet(ctx context.Context) (*repository.Wallet, error) {
	var wallet repository.Wallet
	walletID := uuid.New()
	err := r.pool.QueryRow(ctx, "INSERT INTO wallets (id, balance) VALUES ($1, 0) RETURNING id, balance", walletID).Scan(&wallet.ID, &wallet.Balance)
	var pgErr *pgconn.PgError
	if err != nil {
		if stderrors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, apperrors.ErrWalletAlreadyExists
		}
		return nil, apperrors.NewDatabaseError("создании кошелька", err)
	}
	return &wallet, nil
}
