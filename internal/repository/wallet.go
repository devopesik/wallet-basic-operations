package repository

import (
	"context"

	"github.com/google/uuid"
)

type WalletRepository interface {
	GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error)
	UpdateBalance(ctx context.Context, walletID uuid.UUID, amount int64, isDeposit bool) error
	CreateWallet(ctx context.Context, walletID uuid.UUID) error
}
