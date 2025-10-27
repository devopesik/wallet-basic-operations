package service

import (
	"context"

	"github.com/google/uuid"
)

type OperationType string

const (
	OperationDeposit  OperationType = "DEPOSIT"
	OperationWithdraw OperationType = "WITHDRAW"
)

type WalletService interface {
	ProcessOperation(ctx context.Context, walletID uuid.UUID, opType OperationType, amount int64) error
	GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error)
	CreateWallet(ctx context.Context, walletID uuid.UUID) error
}
