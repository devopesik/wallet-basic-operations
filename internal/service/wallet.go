package service

import (
	"context"
	"fmt"

	"github.com/devopesik/wallet-basic-operations/internal/repository"
	"github.com/google/uuid"
)

type walletService struct {
	repo repository.WalletRepository
}

func NewWalletService(repo repository.WalletRepository) WalletService {
	return &walletService{repo: repo}
}

func (s *walletService) ProcessOperation(ctx context.Context, walletID uuid.UUID, opType OperationType, amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("сумма должна быть положительной")
	}

	switch opType {
	case OperationDeposit:
		return s.repo.UpdateBalance(ctx, walletID, amount, true)
	case OperationWithdraw:
		return s.repo.UpdateBalance(ctx, walletID, amount, false)
	default:
		return fmt.Errorf("неизвестный тип операции: %s", opType)
	}
}

func (s *walletService) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	return s.repo.GetBalance(ctx, walletID)
}

func (s *walletService) CreateWallet(ctx context.Context, walletID uuid.UUID) error {
	return s.repo.CreateWallet(ctx, walletID)
}
