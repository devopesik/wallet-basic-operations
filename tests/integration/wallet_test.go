package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestWalletIntegration(t *testing.T) {
	baseURL, cleanup := testServer(t)
	defer cleanup()

	// 0. Создание кошелька (UUID генерируется автоматически)
	resp, err := http.Post(baseURL+"/api/v1/wallets", "application/json", bytes.NewReader([]byte("{}")))
	if err != nil {
		t.Fatalf("ошибка при создании кошелька: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("ожидался статус 201, получен %d, тело: %s", resp.StatusCode, string(body))
	}

	var createResp struct {
		WalletId string `json:"walletId"`
		Balance  int64  `json:"balance"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		t.Fatalf("ошибка декодирования ответа создания кошелька: %v", err)
	}
	_ = resp.Body.Close()

	walletID := createResp.WalletId
	if walletID == "" {
		t.Fatal("UUID не был возвращен при создании кошелька")
	}
	if createResp.Balance != 0 {
		t.Errorf("ожидался баланс 0, получен %d", createResp.Balance)
	}

	// 1. Депозит
	depositReq := map[string]interface{}{
		"walletId":      walletID,
		"operationType": "DEPOSIT",
		"amount":        1000,
	}
	depositBody, _ := json.Marshal(depositReq)
	resp, err = http.Post(baseURL+"/api/v1/wallet", "application/json", bytes.NewReader(depositBody))
	if err != nil {
		t.Fatalf("ошибка при депозите: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("ожидался статус 204, получен %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// 2. Получение баланса
	resp, err = http.Get(baseURL + "/api/v1/wallets/" + walletID)
	if err != nil {
		t.Fatalf("ошибка при получении баланса: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("ожидался статус 200, получен %d", resp.StatusCode)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var balanceResp struct {
		WalletId string `json:"walletId"`
		Balance  int64  `json:"balance"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&balanceResp); err != nil {
		t.Fatalf("ошибка декодирования ответа: %v", err)
	}
	if balanceResp.Balance != 1000 {
		t.Errorf("ожидался баланс 1000, получен %d", balanceResp.Balance)
	}

	// 3. Списание
	withdrawReq := map[string]interface{}{
		"walletId":      walletID,
		"operationType": "WITHDRAW",
		"amount":        300,
	}
	withdrawBody, _ := json.Marshal(withdrawReq)
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/wallet", bytes.NewReader(withdrawBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("ошибка при списании: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("ожидался статус 204, получен %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// 4. Проверка итогового баланса
	resp, err = http.Get(baseURL + "/api/v1/wallets/" + walletID)
	if err != nil {
		t.Fatalf("ошибка при финальном запросе баланса: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("ошибка: статус %d, тело: %s", resp.StatusCode, string(errorBody))
	}
	if err := json.NewDecoder(resp.Body).Decode(&balanceResp); err != nil {
		t.Fatalf("ошибка декодирования: %v", err)
	}
	if balanceResp.Balance != 700 {
		t.Errorf("ожидался баланс 700, получен %d", balanceResp.Balance)
	}
}
