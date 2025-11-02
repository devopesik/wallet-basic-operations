package service

import (
	"context"

	"github.com/devopesik/wallet-basic-operations/internal/repository"
	"github.com/google/uuid"
)

// OperationType представляет тип операции
type OperationType string

const (
	OperationDeposit  OperationType = "DEPOSIT"
	OperationWithdraw OperationType = "WITHDRAW"
)

type WalletService interface {
	Deposit(ctx context.Context, walletID uuid.UUID, amount int64) error
	Withdraw(ctx context.Context, walletID uuid.UUID, amount int64) error
	GetWallet(ctx context.Context, walletID uuid.UUID) (*repository.Wallet, error)
	CreateWallet(ctx context.Context) (*repository.Wallet, error)
}
