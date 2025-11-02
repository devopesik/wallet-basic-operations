package handler

import (
	"encoding/json"
	"log"
	"net/http"

	apperrors "github.com/devopesik/wallet-basic-operations/internal/errors"
	"github.com/devopesik/wallet-basic-operations/internal/generated"
	"github.com/devopesik/wallet-basic-operations/internal/service"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type walletHandler struct {
	service service.WalletService
}

func NewWalletHandler(svc service.WalletService) generated.ServerInterface {
	return &walletHandler{service: svc}
}

func (h *walletHandler) ProcessWalletOperation(w http.ResponseWriter, r *http.Request) {
	req, err := validateWalletOperationRequest(r)
	if err != nil {
		handleError(w, err)
		return
	}

	walletID, err := validateWalletID(req.WalletId)
	if err != nil {
		handleError(w, err)
		return
	}

	if err := validateAmount(req.Amount); err != nil {
		handleError(w, err)
		return
	}

	if err := validateOperationType(req.OperationType); err != nil {
		handleError(w, err)
		return
	}

	switch req.OperationType {
	case generated.DEPOSIT:
		err = h.service.Deposit(r.Context(), walletID, req.Amount)
	case generated.WITHDRAW:
		err = h.service.Withdraw(r.Context(), walletID, req.Amount)
	}

	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *walletHandler) GetWalletBalance(w http.ResponseWriter, r *http.Request, walletId openapi_types.UUID) {
	walletID, err := validateWalletID(walletId)
	if err != nil {
		handleError(w, err)
		return
	}

	wallet, err := h.service.GetWallet(r.Context(), walletID)
	if err != nil {
		handleError(w, err)
		return
	}

	resp := generated.WalletBalanceResponse{
		WalletId: &walletId,
		Balance:  &wallet.Balance,
	}
	writeJSON(w, resp, http.StatusOK)
}

func (h *walletHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	wallet, err := h.service.CreateWallet(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}

	// Конвертируем uuid.UUID в openapi_types.UUID для ответа
	walletIdResponse := openapi_types.UUID(wallet.ID)
	resp := generated.WalletBalanceResponse{
		WalletId: &walletIdResponse,
		Balance:  &wallet.Balance,
	}
	writeJSON(w, resp, http.StatusCreated)
}

func (h *walletHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	resp := struct {
		Status string `json:"status"`
	}{
		Status: status,
	}
	writeJSON(w, resp, http.StatusOK)
}

func writeJSON(w http.ResponseWriter, v interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("could not write json response: %v", err)
	}
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	errResp := generated.Error{Message: &message}
	writeJSON(w, errResp, status)
}

// validateWalletOperationRequest валидирует и декодирует запрос на операцию с кошельком
func validateWalletOperationRequest(r *http.Request) (*generated.WalletOperationRequest, error) {
	// Ограничиваем размер тела запроса для защиты от больших запросов (1MB)
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	defer r.Body.Close()

	var req generated.WalletOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, apperrors.ErrInvalidJSON
	}
	return &req, nil
}

// validateWalletID валидирует и конвертирует openapi_types.UUID в uuid.UUID
func validateWalletID(walletId openapi_types.UUID) (uuid.UUID, error) {
	walletID, err := uuid.Parse(walletId.String())
	if err != nil {
		return uuid.Nil, apperrors.ErrInvalidWalletID
	}
	return walletID, nil
}

// validateAmount валидирует сумму операции
func validateAmount(amount int64) error {
	if amount <= 0 {
		return apperrors.ErrInvalidAmount
	}
	return nil
}

// validateOperationType валидирует тип операции
func validateOperationType(opType generated.WalletOperationRequestOperationType) error {
	switch opType {
	case generated.DEPOSIT, generated.WITHDRAW:
		return nil
	default:
		return apperrors.ErrInvalidOperationType
	}
}

// handleError обрабатывает ошибку и отправляет соответствующий HTTP ответ
func handleError(w http.ResponseWriter, err error) {
	appErr, ok := apperrors.AsAppError(err)
	if !ok {
		// Если это не AppError, логируем как внутреннюю ошибку и возвращаем общий ответ
		log.Printf("internal error: %v", err)
		message := "внутренняя ошибка"
		writeJSONError(w, message, http.StatusInternalServerError)
		return
	}

	// Логируем ошибки с контекстом
	statusCode := appErr.HTTPStatus()
	if statusCode >= 500 {
		// Внутренние ошибки (5xx) логируем с полным контекстом
		log.Printf("server error [%d]: %s: %v", statusCode, appErr.Message, appErr.Err)
	} else {
		// Клиентские ошибки (4xx) логируем на уровне предупреждения
		log.Printf("client error [%d]: %s", statusCode, appErr.Message)
	}

	writeJSONError(w, appErr.Message, statusCode)
}
