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
	if amount <= 0 {
		return apperrors.ErrInvalidAmount
	}

	query := "UPDATE wallets SET balance = balance + $1 WHERE id = $2"
	result, err := r.pool.Exec(ctx, query, amount, walletID)
	if err != nil {
		return apperrors.NewDatabaseError("пополнении баланса", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.ErrWalletNotFound
	}

	return nil
}

func (r *walletRepository) Withdraw(ctx context.Context, walletID uuid.UUID, amount int64) error {
	if amount <= 0 {
		return apperrors.ErrInvalidAmount
	}

	query := "UPDATE wallets SET balance = balance - $1 WHERE id = $2 AND balance >= $1"
	result, err := r.pool.Exec(ctx, query, amount, walletID)
	if err != nil {
		return apperrors.NewDatabaseError("списании баланса", err)
	}

	if result.RowsAffected() == 0 {
		// Проверяем, существует ли кошелек
		var count int64
		err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM wallets WHERE id = $1", walletID).Scan(&count)
		if err != nil {
			return apperrors.NewDatabaseError("проверке кошелька", err)
		}
		if count == 0 {
			return apperrors.ErrWalletNotFound
		}
		return apperrors.ErrInsufficientFunds
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
