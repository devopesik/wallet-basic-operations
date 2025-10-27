package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockWalletRepository struct {
	mock.Mock
}

func (m *MockWalletRepository) CreateWallet(ctx context.Context, walletID uuid.UUID) error {
	args := m.Called(ctx, walletID)
	return args.Error(0)
}

func (m *MockWalletRepository) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	args := m.Called(ctx, walletID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockWalletRepository) UpdateBalance(ctx context.Context, walletID uuid.UUID, amount int64, isDeposit bool) error {
	args := m.Called(ctx, walletID, amount, isDeposit)
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

func TestWalletService_ProcessOperation_InvalidAmount(t *testing.T) {
	repo := new(MockWalletRepository)
	svc := NewWalletService(repo)

	cases := []int64{0, -1, -1000}
	for _, amount := range cases {
		t.Run("amount="+string(rune(amount)), func(t *testing.T) {
			err := svc.ProcessOperation(context.Background(), testWalletID, OperationDeposit, amount)
			if err == nil {
				t.Fatal("ожидалась ошибка при недопустимой сумме")
			}
			if err.Error() != "сумма должна быть положительной" {
				t.Errorf("некорректное сообщение об ошибке: %v", err)
			}
		})
	}
}

func TestWalletService_ProcessOperation_UnknownOperationType(t *testing.T) {
	repo := new(MockWalletRepository)
	svc := NewWalletService(repo)

	err := svc.ProcessOperation(context.Background(), testWalletID, "INVALID", 100)
	if err == nil {
		t.Fatal("ожидалась ошибка при неизвестном типе операции")
	}
	if err.Error() != "неизвестный тип операции: INVALID" {
		t.Errorf("некорректное сообщение об ошибке: %v", err)
	}
}

func TestWalletService_ProcessOperation_Deposit_Success(t *testing.T) {
	repo := new(MockWalletRepository)
	repo.On("UpdateBalance", mock.Anything, testWalletID, int64(500), true).Return(nil)
	svc := NewWalletService(repo)

	err := svc.ProcessOperation(context.Background(), testWalletID, OperationDeposit, 500)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	repo.AssertExpectations(t)
}

func TestWalletService_ProcessOperation_Deposit_WalletNotFound(t *testing.T) {
	repo := new(MockWalletRepository)
	repo.On("UpdateBalance", mock.Anything, testWalletID, int64(100), true).Return(errors.New("кошелёк не найден"))
	svc := NewWalletService(repo)

	err := svc.ProcessOperation(context.Background(), testWalletID, OperationDeposit, 100)
	if err == nil {
		t.Fatal("ожидалась ошибка 'кошелёк не найден'")
	}
	if err.Error() != "кошелёк не найден" {
		t.Errorf("некорректное сообщение: %v", err)
	}

	repo.AssertExpectations(t)
}

func TestWalletService_ProcessOperation_Withdraw_Success(t *testing.T) {
	repo := new(MockWalletRepository)
	repo.On("UpdateBalance", mock.Anything, testWalletID, int64(200), false).Return(nil)
	svc := NewWalletService(repo)

	err := svc.ProcessOperation(context.Background(), testWalletID, OperationWithdraw, 200)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	repo.AssertExpectations(t)
}

func TestWalletService_ProcessOperation_Withdraw_InsufficientFunds(t *testing.T) {
	repo := new(MockWalletRepository)
	repo.On("UpdateBalance", mock.Anything, testWalletID, int64(1000), false).Return(errors.New("недостаточно средств"))
	svc := NewWalletService(repo)

	err := svc.ProcessOperation(context.Background(), testWalletID, OperationWithdraw, 1000)
	if err == nil {
		t.Fatal("ожидалась ошибка 'недостаточно средств'")
	}
	if err.Error() != "недостаточно средств" {
		t.Errorf("некорректное сообщение: %v", err)
	}

	repo.AssertExpectations(t)
}

func TestWalletService_GetBalance_Success(t *testing.T) {
	repo := new(MockWalletRepository)
	repo.On("GetBalance", mock.Anything, testWalletID).Return(int64(750), nil)
	svc := NewWalletService(repo)

	balance, err := svc.GetBalance(context.Background(), testWalletID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if balance != 750 {
		t.Errorf("ожидался баланс 750, получен %d", balance)
	}

	repo.AssertExpectations(t)
}

func TestWalletService_GetBalance_NotFound(t *testing.T) {
	repo := new(MockWalletRepository)
	repo.On("GetBalance", mock.Anything, testWalletID).Return(int64(0), errors.New("кошелёк не найден"))
	svc := NewWalletService(repo)

	_, err := svc.GetBalance(context.Background(), testWalletID)
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
	repo.On("CreateWallet", mock.Anything, testWalletID).Return(nil)
	svc := NewWalletService(repo)

	err := svc.CreateWallet(context.Background(), testWalletID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	repo.AssertExpectations(t)
}

func TestWalletService_CreateWallet_AlreadyExists(t *testing.T) {
	repo := new(MockWalletRepository)
	repo.On("CreateWallet", mock.Anything, testWalletID).Return(errors.New("кошелёк уже существует"))
	svc := NewWalletService(repo)

	err := svc.CreateWallet(context.Background(), testWalletID)
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
	repo.On("CreateWallet", mock.Anything, testWalletID).Return(errors.New("ошибка подключения к базе данных"))
	svc := NewWalletService(repo)

	err := svc.CreateWallet(context.Background(), testWalletID)
	if err == nil {
		t.Fatal("ожидалась ошибка от репозитория")
	}
	if err.Error() != "ошибка подключения к базе данных" {
		t.Errorf("некорректное сообщение: %v", err)
	}

	repo.AssertExpectations(t)
}
