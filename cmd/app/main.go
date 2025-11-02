package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/devopesik/wallet-basic-operations/internal/app"
	"github.com/devopesik/wallet-basic-operations/internal/config"
)

const cfgPath = "config.env"
const shutdownTimeout = 10 * time.Second

func main() {
	cfg := config.Load(cfgPath)
	application, err := app.StartServer(cfg)
	if err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}

	// Канал для получения сигналов ОС
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Ожидание сигнала завершения
	sig := <-sigChan
	log.Printf("Получен сигнал %v, начинаю graceful shutdown...", sig)

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Останавливаем приложение (сервер и пул БД)
	if err := application.Shutdown(ctx); err != nil {
		log.Printf("Ошибка при остановке приложения: %v", err)
		// Принудительно закрываем сервер, если graceful shutdown не удался
		if err := application.Server.Close(); err != nil {
			log.Fatalf("Не удалось закрыть сервер: %v", err)
		}
	} else {
		log.Println("Приложение корректно остановлено")
	}
}
