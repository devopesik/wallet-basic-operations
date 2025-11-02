package integration

import (
	"context"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/devopesik/wallet-basic-operations/internal/app"
	"github.com/devopesik/wallet-basic-operations/internal/config"
)

func testServer(t *testing.T) (string, func()) {

	cfg := config.Load("../../config.env")
	if dbHost, ok := os.LookupEnv("DB_HOST"); ok {
		cfg.DBHost = dbHost
	}

	application, err := app.StartServer(cfg)
	if err != nil {
		t.Fatalf("не удалось запустить сервер: %v", err)
	}

	baseURL := "http://localhost:" + cfg.ServerPort

	const timeout = 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	log.Println("Waiting for server to become healthy...")
	healthy := false
	for {
		// Проверяем, не истек ли наш таймаут
		if ctx.Err() != nil {
			t.Fatalf("Сервер не стал доступен в течение %.0f секунд. Логи сервера могут содержать ошибку.", timeout.Seconds())
		}

		// Пытаемся сделать запрос к /health
		resp, err := http.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			healthy = true
			_ = resp.Body.Close()
			break // Сервер готов, выходим из цикла
		}

		// Если не получилось, ждем немного и пробуем снова
		time.Sleep(200 * time.Millisecond)
	}

	if !healthy {
		t.Fatal("Сервер не смог запуститься корректно.")
	}

	log.Println("Server is healthy, starting tests.")

	cleanup := func() {
		if err := application.Shutdown(context.Background()); err != nil {
			log.Printf("ошибка при остановке приложения: %v", err)
		}
	}

	return baseURL, cleanup
}
