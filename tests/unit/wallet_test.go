package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/devopesik/wallet-basic-operations/internal/repository"
	"github.com/devopesik/wallet-basic-operations/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockWalletRepository struct {
	mock.Mock
}

func (m *MockWalletRepository) CreateWallet(ctx context.Context) (*repository.Wallet, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Wallet), args.Error(1)
}

func (m *MockWalletRepository) GetWallet(ctx context.Context, walletID uuid.UUID) (*repository.Wallet, error) {
	args := m.Called(ctx, walletID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Wallet), args.Error(1)
}

func (m *MockWalletRepository) Deposit(ctx context.Context, walletID uuid.UUID, amount int64) error {
	args := m.Called(ctx, walletID, amount)
	return args.Error(0)
}

func (m *MockWalletRepository) Withdraw(ctx context.Context, walletID uuid.UUID, amount int64) error {
	args := m.Called(ctx, walletID, amount)
	return args.Error(0)
}

func mustUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}

var testWalletID = mustUUID("123e4567-e89b-12d3-a456-426614174000")

func TestWalletService_Deposit_InvalidAmount(t *testing.T) {
	repo := new(MockWalletRepository)
	svc := service.NewWalletService(repo)

	cases := []int64{0, -1, -1000}
	for _, amount := range cases {
		t.Run("amount="+string(rune(amount)), func(t *testing.T) {
			err := svc.Deposit(context.Background(), testWalletID, amount)
			if err == nil {
				t.Fatal("ожидалась ошибка при недопустимой сумме")
			}
			if err.Error() != "сумма должна быть положительной" {
				t.Errorf("некорректное сообщение об ошибке: %v", err)
			}
		})
	}
}

func TestWalletService_Deposit_Success(t *testing.T) {
	repo := new(MockWalletRepository)
	repo.On("Deposit", mock.Anything, testWalletID, int64(500)).Return(nil)
	svc := service.NewWalletService(repo)

	err := svc.Deposit(context.Background(), testWalletID, 500)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	repo.AssertExpectations(t)
}

func TestWalletService_Deposit_WalletNotFound(t *testing.T) {
	repo := new(MockWalletRepository)
	repo.On("Deposit", mock.Anything, testWalletID, int64(100)).Return(errors.New("кошелёк не найден"))
	svc := service.NewWalletService(repo)

	err := svc.Deposit(context.Background(), testWalletID, 100)
	if err == nil {
		t.Fatal("ожидалась ошибка 'кошелёк не найден'")
	}
	if err.Error() != "кошелёк не найден" {
		t.Errorf("некорректное сообщение: %v", err)
	}

	repo.AssertExpectations(t)
}

func TestWalletService_Withdraw_InvalidAmount(t *testing.T) {
	repo := new(MockWalletRepository)
	svc := service.NewWalletService(repo)

	cases := []int64{0, -1, -1000}
	for _, amount := range cases {
		t.Run("amount="+string(rune(amount)), func(t *testing.T) {
			err := svc.Withdraw(context.Background(), testWalletID, amount)
			if err == nil {
				t.Fatal("ожидалась ошибка при недопустимой сумме")
			}
			if err.Error() != "сумма должна быть положительной" {
				t.Errorf("некорректное сообщение об ошибке: %v", err)
			}
		})
	}
}

func TestWalletService_Withdraw_Success(t *testing.T) {
	repo := new(MockWalletRepository)
	repo.On("Withdraw", mock.Anything, testWalletID, int64(200)).Return(nil)
	svc := service.NewWalletService(repo)

	err := svc.Withdraw(context.Background(), testWalletID, 200)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	repo.AssertExpectations(t)
}

func TestWalletService_Withdraw_InsufficientFunds(t *testing.T) {
	repo := new(MockWalletRepository)
	repo.On("Withdraw", mock.Anything, testWalletID, int64(1000)).Return(errors.New("недостаточно средств"))
	svc := service.NewWalletService(repo)

	err := svc.Withdraw(context.Background(), testWalletID, 1000)
	if err == nil {
		t.Fatal("ожидалась ошибка 'недостаточно средств'")
	}
	if err.Error() != "недостаточно средств" {
		t.Errorf("некорректное сообщение: %v", err)
	}

	repo.AssertExpectations(t)
}

func TestWalletService_Withdraw_WalletNotFound(t *testing.T) {
	repo := new(MockWalletRepository)
	repo.On("Withdraw", mock.Anything, testWalletID, int64(100)).Return(errors.New("кошелёк не найден"))
	svc := service.NewWalletService(repo)

	err := svc.Withdraw(context.Background(), testWalletID, 100)
	if err == nil {
		t.Fatal("ожидалась ошибка 'кошелёк не найден'")
	}
	if err.Error() != "кошелёк не найден" {
		t.Errorf("некорректное сообщение: %v", err)
	}

	repo.AssertExpectations(t)
}

func TestWalletService_GetWallet_Success(t *testing.T) {
	repo := new(MockWalletRepository)
	expectedWallet := &repository.Wallet{
		ID:      testWalletID,
		Balance: 750,
	}
	repo.On("GetWallet", mock.Anything, testWalletID).Return(expectedWallet, nil)
	svc := service.NewWalletService(repo)

	wallet, err := svc.GetWallet(context.Background(), testWalletID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if wallet.Balance != 750 {
		t.Errorf("ожидался баланс 750, получен %d", wallet.Balance)
	}
	if wallet.ID != testWalletID {
		t.Errorf("ожидался ID %v, получен %v", testWalletID, wallet.ID)
	}

	repo.AssertExpectations(t)
}

func TestWalletService_GetWallet_NotFound(t *testing.T) {
	repo := new(MockWalletRepository)
	repo.On("GetWallet", mock.Anything, testWalletID).Return(nil, errors.New("кошелёк не найден"))
	svc := service.NewWalletService(repo)

	_, err := svc.GetWallet(context.Background(), testWalletID)
	if err == nil {
		t.Fatal("ожидалась ошибка 'кошелёк не найден'")
	}
	if err.Error() != "кошелёк не найден" {
		t.Errorf("некорректное сообщение: %v", err)
	}

	repo.AssertExpectations(t)
}

func TestWalletService_CreateWallet_Success(t *testing.T) {
	repo := new(MockWalletRepository)
	expectedWallet := &repository.Wallet{
		ID:      testWalletID,
		Balance: 0,
	}
	repo.On("CreateWallet", mock.Anything).Return(expectedWallet, nil)
	svc := service.NewWalletService(repo)

	wallet, err := svc.CreateWallet(context.Background())
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if wallet.ID != testWalletID {
		t.Errorf("ожидался ID %v, получен %v", testWalletID, wallet.ID)
	}
	if wallet.Balance != 0 {
		t.Errorf("ожидался баланс 0, получен %d", wallet.Balance)
	}

	repo.AssertExpectations(t)
}

func TestWalletService_CreateWallet_AlreadyExists(t *testing.T) {
	repo := new(MockWalletRepository)
	repo.On("CreateWallet", mock.Anything).Return(nil, errors.New("кошелёк уже существует"))
	svc := service.NewWalletService(repo)

	_, err := svc.CreateWallet(context.Background())
	if err == nil {
		t.Fatal("ожидалась ошибка 'кошелёк уже существует'")
	}
	if err.Error() != "кошелёк уже существует" {
		t.Errorf("некорректное сообщение: %v", err)
	}

	repo.AssertExpectations(t)
}

func TestWalletService_CreateWallet_RepositoryError(t *testing.T) {
	repo := new(MockWalletRepository)
	repo.On("CreateWallet", mock.Anything).Return(nil, errors.New("ошибка подключения к базе данных"))
	svc := service.NewWalletService(repo)

	_, err := svc.CreateWallet(context.Background())
	if err == nil {
		t.Fatal("ожидалась ошибка от репозитория")
	}
	if err.Error() != "ошибка подключения к базе данных" {
		t.Errorf("некорректное сообщение: %v", err)
	}

	repo.AssertExpectations(t)
}
