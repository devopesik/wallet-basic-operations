package repository

import (
	"context"

	"github.com/google/uuid"
)

// Wallet представляет структуру кошелька
type Wallet struct {
	ID      uuid.UUID
	Balance int64
}

type WalletRepository interface {
	GetWallet(ctx context.Context, walletID uuid.UUID) (*Wallet, error)
	Deposit(ctx context.Context, walletID uuid.UUID, amount int64) error
	Withdraw(ctx context.Context, walletID uuid.UUID, amount int64) error
	CreateWallet(ctx context.Context) (*Wallet, error)
}
