package errors

import (
	"fmt"
	"net/http"
)

// AppError представляет базовую структуру ошибки приложения
type AppError struct {
	Code       int    // HTTP код статуса
	Message    string // Сообщение ошибки
	Err        error  // Оригинальная ошибка (опционально)
	StatusCode int    // HTTP статус код для ответа
}

// Error реализует интерфейс error
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap возвращает оригинальную ошибку для использования с errors.Is и errors.As
func (e *AppError) Unwrap() error {
	return e.Err
}

// HTTPStatus возвращает HTTP статус код ошибки
func (e *AppError) HTTPStatus() int {
	return e.StatusCode
}

// ErrWalletNotFound - кошелёк не найден
var ErrWalletNotFound = &AppError{
	Code:       ErrorCodeWalletNotFound,
	Message:    "кошелёк не найден",
	StatusCode: http.StatusNotFound,
}

// ErrInsufficientFunds - недостаточно средств
var ErrInsufficientFunds = &AppError{
	Code:       ErrorCodeInsufficientFunds,
	Message:    "недостаточно средств",
	StatusCode: http.StatusConflict,
}

// ErrInvalidAmount - некорректная сумма
var ErrInvalidAmount = &AppError{
	Code:       ErrorCodeInvalidAmount,
	Message:    "сумма должна быть положительной",
	StatusCode: http.StatusBadRequest,
}

// ErrInvalidOperationType - некорректный тип операции
var ErrInvalidOperationType = &AppError{
	Code:       ErrorCodeInvalidOperationType,
	Message:    "недопустимый operationType",
	StatusCode: http.StatusBadRequest,
}

// ErrWalletAlreadyExists - кошелёк уже существует
var ErrWalletAlreadyExists = &AppError{
	Code:       ErrorCodeWalletAlreadyExists,
	Message:    "кошелёк уже существует",
	StatusCode: http.StatusConflict,
}

// ErrInvalidJSON - некорректный JSON
var ErrInvalidJSON = &AppError{
	Code:       ErrorCodeInvalidJSON,
	Message:    "некорректный JSON",
	StatusCode: http.StatusBadRequest,
}

// ErrInvalidWalletID - некорректный walletId
var ErrInvalidWalletID = &AppError{
	Code:       ErrorCodeInvalidWalletID,
	Message:    "некорректный walletId",
	StatusCode: http.StatusBadRequest,
}

// ErrDatabaseError - ошибка базы данных
var ErrDatabaseError = &AppError{
	Code:       ErrorCodeDatabaseError,
	Message:    "внутренняя ошибка",
	StatusCode: http.StatusInternalServerError,
}

// Коды ошибок
const (
	ErrorCodeWalletNotFound       = 1001
	ErrorCodeInsufficientFunds    = 1002
	ErrorCodeInvalidAmount        = 1003
	ErrorCodeInvalidOperationType = 1004
	ErrorCodeWalletAlreadyExists  = 1005
	ErrorCodeInvalidJSON          = 1006
	ErrorCodeInvalidWalletID      = 1007
	ErrorCodeDatabaseError        = 2001
)

// Вспомогательные функции для создания ошибок с контекстом

// NewWalletNotFound возвращает ошибку с контекстом
func NewWalletNotFound(err error) *AppError {
	return &AppError{
		Code:       ErrorCodeWalletNotFound,
		Message:    ErrWalletNotFound.Message,
		Err:        err,
		StatusCode: ErrWalletNotFound.StatusCode,
	}
}

// NewInsufficientFunds возвращает ошибку с контекстом
func NewInsufficientFunds() *AppError {
	return &AppError{
		Code:       ErrorCodeInsufficientFunds,
		Message:    ErrInsufficientFunds.Message,
		StatusCode: ErrInsufficientFunds.StatusCode,
	}
}

// NewInvalidAmount возвращает ошибку с контекстом
func NewInvalidAmount(err error) *AppError {
	return &AppError{
		Code:       ErrorCodeInvalidAmount,
		Message:    ErrInvalidAmount.Message,
		Err:        err,
		StatusCode: ErrInvalidAmount.StatusCode,
	}
}

// NewInvalidOperationType возвращает ошибку с контекстом
func NewInvalidOperationType(opType string) *AppError {
	return &AppError{
		Code:       ErrorCodeInvalidOperationType,
		Message:    fmt.Sprintf("%s: %s", ErrInvalidOperationType.Message, opType),
		StatusCode: ErrInvalidOperationType.StatusCode,
	}
}

// NewWalletAlreadyExists возвращает ошибку с контекстом
func NewWalletAlreadyExists(err error) *AppError {
	return &AppError{
		Code:       ErrorCodeWalletAlreadyExists,
		Message:    ErrWalletAlreadyExists.Message,
		Err:        err,
		StatusCode: ErrWalletAlreadyExists.StatusCode,
	}
}

// NewDatabaseError возвращает ошибку базы данных с контекстом
func NewDatabaseError(operation string, err error) *AppError {
	return &AppError{
		Code:       ErrorCodeDatabaseError,
		Message:    fmt.Sprintf("ошибка базы данных при %s: %v", operation, err),
		Err:        err,
		StatusCode: ErrDatabaseError.StatusCode,
	}
}

// IsAppError проверяет, является ли ошибка AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// AsAppError преобразует ошибку в AppError
func AsAppError(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}

	// Пробуем получить AppError напрямую
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}

	// Пробуем найти AppError в цепочке ошибок
	if unwrapper, ok := err.(interface{ Unwrap() error }); ok {
		if unwrapped, ok := AsAppError(unwrapper.Unwrap()); ok {
			return unwrapped, true
		}
	}

	return nil, false
}
