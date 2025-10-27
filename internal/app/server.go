package app

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/devopesik/wallet-basic-operations/internal/config"
	"github.com/devopesik/wallet-basic-operations/internal/generated"
	"github.com/devopesik/wallet-basic-operations/internal/handlers"
	"github.com/devopesik/wallet-basic-operations/internal/repository/postgres"
	"github.com/devopesik/wallet-basic-operations/internal/service"
	"github.com/go-chi/chi/v5"
)

func StartServer(cfg *config.Config) (*http.Server, error) {
	if err := postgres.RunMigrations(cfg); err != nil {
		return nil, err
	}

	pool, err := postgres.NewPool(cfg)
	if err != nil {
		return nil, err
	}

	repo := postgres.NewWalletRepository(pool)
	svc := service.NewWalletService(repo)
	hdl := handler.NewWalletHandler(svc)

	r := chi.NewRouter()
	generated.HandlerFromMux(hdl, r)

	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		BaseContext: func(l net.Listener) context.Context {
			return context.Background()
		},
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Сервер завершил работу с ошибкой: %v", err)
		}
	}()

	return server, nil
}
