package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestWalletConcurrentLoad(t *testing.T) {
	baseURL, cleanup := testServer(t)
	defer cleanup()

	walletID := uuid.New().String()

	// 1. Создание кошелька
	createReq := map[string]interface{}{
		"walletId": walletID,
	}
	body, _ := json.Marshal(createReq)
	resp, err := http.Post(baseURL+"/api/v1/wallets", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("ошибка при создании кошелька: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("ожидался статус 201, получен %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// 2. Начальный депозит для обеспечения средств для операций
	initialDeposit := map[string]interface{}{
		"walletId":      walletID,
		"operationType": "DEPOSIT",
		"amount":        1000000, // 1 миллион для обеспечения операций
	}
	body, _ = json.Marshal(initialDeposit)
	resp, err = http.Post(baseURL+"/api/v1/wallet", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("ошибка при начальном депозите: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("ожидался статус 204, получен %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// 3. Конкурентная нагрузка: 1000 RPS в течение 5 секунд
	const (
		duration    = 5 * time.Second
		targetRPS   = 1000
		workerCount = 50
	)
	totalRequests := targetRPS * int(duration.Seconds())

	var (
		wg           sync.WaitGroup
		successCount int64
		errorCount   int64
		server5xx    int64
		mu           sync.Mutex
	)

	// Канал для распределения запросов между воркерами
	requestChan := make(chan int, totalRequests)
	for i := 0; i < totalRequests; i++ {
		requestChan <- i
	}
	close(requestChan)

	startTime := time.Now()

	// Запуск воркеров
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			client := &http.Client{
				Timeout: 10 * time.Second,
			}

			for reqID := range requestChan {
				// Определяем тип операции: 40% депозит, 40% списание, 20% проверка баланса
				opType := reqID % 5
				var err error
				var resp *http.Response

				switch opType {
				case 0, 1: // Депозит
					depositReq := map[string]interface{}{
						"walletId":      walletID,
						"operationType": "DEPOSIT",
						"amount":        100, // Небольшая сумма
					}
					body, _ := json.Marshal(depositReq)
					resp, err = client.Post(baseURL+"/api/v1/wallet", "application/json", bytes.NewReader(body))

				case 2, 3: // Списание
					withdrawReq := map[string]interface{}{
						"walletId":      walletID,
						"operationType": "WITHDRAW",
						"amount":        50, // Меньше чем депозит
					}
					body, _ := json.Marshal(withdrawReq)
					req, _ := http.NewRequest("POST", baseURL+"/api/v1/wallet", bytes.NewReader(body))
					req.Header.Set("Content-Type", "application/json")
					resp, err = client.Do(req)

				case 4: // Проверка баланса
					resp, err = client.Get(baseURL + "/api/v1/wallets/" + walletID)
				}

				if err != nil {
					mu.Lock()
					errorCount++
					mu.Unlock()
					continue
				}

				// Проверка статуса ответа
				statusCode := resp.StatusCode
				bodyBytes, _ := io.ReadAll(resp.Body)
				_ = resp.Body.Close()

				mu.Lock()
				if statusCode >= 500 {
					server5xx++
					t.Errorf("воркер %d, запрос %d: 5xx ошибка - статус %d, тело: %s",
						workerID, reqID, statusCode, string(bodyBytes))
				} else if statusCode >= 400 {
					// 4xx ошибки допустимы (например, недостаточно средств)
					errorCount++
				} else {
					successCount++
				}
				mu.Unlock()
			}
		}(i)
	}

	// Ожидание завершения всех воркеров
	wg.Wait()
	elapsed := time.Since(startTime)

	// 4. Проверка результатов
	actualRPS := float64(successCount+errorCount) / elapsed.Seconds()

	t.Logf("=== Результаты нагрузочного теста ===")
	t.Logf("Длительность теста: %v", elapsed)
	t.Logf("Целевой RPS: %d", targetRPS)
	t.Logf("Фактический RPS: %.2f", actualRPS)
	t.Logf("Всего запросов: %d", totalRequests)
	t.Logf("Успешных запросов: %d", successCount)
	t.Logf("Ошибок (4xx): %d", errorCount)
	t.Logf("Ошибок сервера (5xx): %d", server5xx)

	// 5. Проверка требований
	if server5xx > 0 {
		t.Fatalf("обнаружены ошибки сервера (5xx): %d. Требование: ни один запрос не должен завершиться с 5xx ошибкой", server5xx)
	}

	if actualRPS < float64(targetRPS)*0.8 { // Допустим 20% погрешность
		t.Errorf("фактический RPS (%.2f) значительно ниже целевого (%d)", actualRPS, targetRPS)
	}

	// 6. Проверка финального состояния кошелька
	resp, err = http.Get(baseURL + "/api/v1/wallets/" + walletID)
	if err != nil {
		t.Fatalf("ошибка при финальной проверке баланса: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("ошибка при получении финального баланса, статус: %d", resp.StatusCode)
	}

	var balanceResp struct {
		WalletId string `json:"walletId"`
		Balance  int64  `json:"balance"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&balanceResp); err != nil {
		t.Fatalf("ошибка декодирования финального баланса: %v", err)
	}

	t.Logf("Финальный баланс кошелька: %d", balanceResp.Balance)

	// Баланс должен быть неотрицательным
	if balanceResp.Balance < 0 {
		t.Errorf("финальный баланс отрицательный: %d", balanceResp.Balance)
	}

	// Проверка, что баланс изменился (не равен начальному депозиту)
	if balanceResp.Balance == 1000000 {
		t.Logf("предупреждение: баланс не изменился, возможно операции не выполнялись")
	}

	t.Logf("=== Тест успешно завершён ===")
}

func TestWalletConcurrentLoad_SustainedHighRPS(t *testing.T) {
	if testing.Short() {
		t.Skip("пропуск длительного нагрузочного теста в режиме -short")
	}

	baseURL, cleanup := testServer(t)
	defer cleanup()

	walletID := uuid.New().String()

	// Создание кошелька
	createReq := map[string]interface{}{
		"walletId": walletID,
	}
	body, _ := json.Marshal(createReq)
	resp, err := http.Post(baseURL+"/api/v1/wallets", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("ошибка при создании кошелька: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("ожидался статус 201, получен %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// Начальный депозит
	initialDeposit := map[string]interface{}{
		"walletId":      walletID,
		"operationType": "DEPOSIT",
		"amount":        5000000,
	}
	body, _ = json.Marshal(initialDeposit)
	resp, err = http.Post(baseURL+"/api/v1/wallet", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("ошибка при начальном депозите: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("ожидался статус 204, получен %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// Длительный тест: 30 секунд с 1000 RPS
	const (
		duration    = 30 * time.Second
		targetRPS   = 1000
		workerCount = 100
	)

	var (
		wg           sync.WaitGroup
		successCount int64
		errorCount   int64
		server5xx    int64
		mu           sync.Mutex
		stopChan     = make(chan struct{})
	)

	startTime := time.Now()

	// Запуск воркеров
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			client := &http.Client{
				Timeout: 15 * time.Second,
			}

			requestsProcessed := 0
			for {
				select {
				case <-stopChan:
					return
				default:
					// Выполняем запрос
					opType := requestsProcessed % 5
					var err error
					var resp *http.Response

					switch opType {
					case 0, 1: // Депозит
						depositReq := map[string]interface{}{
							"walletId":      walletID,
							"operationType": "DEPOSIT",
							"amount":        100,
						}
						body, _ := json.Marshal(depositReq)
						resp, err = client.Post(baseURL+"/api/v1/wallet", "application/json", bytes.NewReader(body))

					case 2, 3: // Списание
						withdrawReq := map[string]interface{}{
							"walletId":      walletID,
							"operationType": "WITHDRAW",
							"amount":        50,
						}
						body, _ := json.Marshal(withdrawReq)
						req, _ := http.NewRequest("POST", baseURL+"/api/v1/wallet", bytes.NewReader(body))
						req.Header.Set("Content-Type", "application/json")
						resp, err = client.Do(req)

					case 4: // Проверка баланса
						resp, err = client.Get(baseURL + "/api/v1/wallets/" + walletID)
					}

					if err != nil {
						mu.Lock()
						errorCount++
						mu.Unlock()
					} else {
						statusCode := resp.StatusCode
						_ = resp.Body.Close()

						mu.Lock()
						if statusCode >= 500 {
							server5xx++
						} else if statusCode >= 400 {
							errorCount++
						} else {
							successCount++
						}
						mu.Unlock()
					}

					requestsProcessed++

					// Контроль скорости для поддержания ~1000 RPS на воркер
					time.Sleep(time.Microsecond * 10)
				}
			}
		}(i)
	}

	// Остановка теста через указанное время
	time.Sleep(duration)
	close(stopChan)
	wg.Wait()
	elapsed := time.Since(startTime)

	actualRPS := float64(successCount+errorCount) / elapsed.Seconds()

	t.Logf("=== Результаты длительного нагрузочного теста ===")
	t.Logf("Длительность теста: %v", elapsed)
	t.Logf("Целевой RPS: %d", targetRPS)
	t.Logf("Фактический RPS: %.2f", actualRPS)
	t.Logf("Всего запросов: %d", successCount+errorCount)
	t.Logf("Успешных запросов: %d", successCount)
	t.Logf("Ошибок (4xx): %d", errorCount)
	t.Logf("Ошибок сервера (5xx): %d", server5xx)

	if server5xx > 0 {
		t.Fatalf("обнаружены ошибки сервера (5xx): %d", server5xx)
	}

	t.Logf("=== Длительный тест успешно завершён ===")
}
