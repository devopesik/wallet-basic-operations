package service

import (
	"context"

	apperrors "github.com/devopesik/wallet-basic-operations/internal/errors"
	"github.com/devopesik/wallet-basic-operations/internal/repository"
	"github.com/google/uuid"
)

type walletService struct {
	repo repository.WalletRepository
}

func NewWalletService(repo repository.WalletRepository) WalletService {
	return &walletService{repo: repo}
}

func (s *walletService) Deposit(ctx context.Context, walletID uuid.UUID, amount int64) error {
	if amount <= 0 {
		return apperrors.ErrInvalidAmount
	}
	return s.repo.Deposit(ctx, walletID, amount)
}

func (s *walletService) Withdraw(ctx context.Context, walletID uuid.UUID, amount int64) error {
	if amount <= 0 {
		return apperrors.ErrInvalidAmount
	}
	return s.repo.Withdraw(ctx, walletID, amount)
}

func (s *walletService) GetWallet(ctx context.Context, walletID uuid.UUID) (*repository.Wallet, error) {
	return s.repo.GetWallet(ctx, walletID)
}

func (s *walletService) CreateWallet(ctx context.Context) (*repository.Wallet, error) {
	return s.repo.CreateWallet(ctx)
}
