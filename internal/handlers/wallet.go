package handler

import (
	"encoding/json"
	"log"
	"net/http"

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
	var req generated.WalletOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "некорректный JSON", http.StatusBadRequest)
		return
	}

	// Конвертируем openapi_types.UUID → uuid.UUID
	walletID, err := uuid.Parse(req.WalletId.String())
	if err != nil {
		http.Error(w, "некорректный walletId", http.StatusBadRequest)
		return
	}

	var opType service.OperationType
	switch req.OperationType {
	case generated.DEPOSIT:
		opType = service.OperationDeposit
	case generated.WITHDRAW:
		opType = service.OperationWithdraw
	default:
		http.Error(w, "недопустимый operationType", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, "amount должен быть положительным", http.StatusBadRequest)
		return
	}

	err = h.service.ProcessOperation(r.Context(), walletID, opType, req.Amount)
	if err != nil {
		switch err.Error() {
		case "кошелёк не найден":
			writeJSONError(w, "кошелёк не найден", http.StatusNotFound)
		case "недостаточно средств":
			writeJSONError(w, "недостаточно средств", http.StatusConflict)
		case "сумма должна быть положительной":
			writeJSONError(w, "amount должен быть положительным", http.StatusBadRequest)
		case "неизвестный тип операции: DEPOSIT", "неизвестный тип операции: WITHDRAW":
			writeJSONError(w, "недопустимый operationType", http.StatusBadRequest)
		default:
			writeJSONError(w, "внутренняя ошибка", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *walletHandler) GetWalletBalance(w http.ResponseWriter, r *http.Request, walletId openapi_types.UUID) {
	walletID, err := uuid.Parse(walletId.String())
	if err != nil {
		writeJSONError(w, "некорректный walletId", http.StatusBadRequest)
		return
	}

	balance, err := h.service.GetBalance(r.Context(), walletID)
	if err != nil {
		switch err.Error() {
		case "кошелёк не найден":
			writeJSONError(w, "кошелёк не найден", http.StatusNotFound)
		default:
			writeJSONError(w, "внутренняя ошибка", http.StatusInternalServerError)
		}
		return
	}

	resp := generated.WalletBalanceResponse{
		WalletId: &walletId,
		Balance:  &balance,
	}
	writeJSON(w, resp, http.StatusOK)
}

func (h *walletHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	var req generated.CreateWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "некорректный JSON", http.StatusBadRequest)
		return
	}

	// Конвертируем openapi_types.UUID → uuid.UUID
	walletID, err := uuid.Parse(req.WalletId.String())
	if err != nil {
		writeJSONError(w, "некорректный walletId", http.StatusBadRequest)
		return
	}

	err = h.service.CreateWallet(r.Context(), walletID)
	if err != nil {
		switch err.Error() {
		case "кошелёк уже существует":
			writeJSONError(w, "кошелёк уже существует", http.StatusConflict)
		default:
			writeJSONError(w, "внутренняя ошибка", http.StatusInternalServerError)
		}
		return
	}

	// Возвращаем созданный кошелёк с балансом 0
	balance := int64(0)
	resp := generated.WalletBalanceResponse{
		WalletId: &req.WalletId,
		Balance:  &balance,
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
		log.Println("could not write json response: %w", err)
	}
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	errResp := generated.Error{Message: &message}
	writeJSON(w, errResp, status)
}
