package main

import (
	"log"

	"github.com/devopesik/wallet-basic-operations/internal/app"
	"github.com/devopesik/wallet-basic-operations/internal/config"
)

const cfgPath = "config.env"

func main() {
	cfg := config.Load(cfgPath)
	_, err := app.StartServer(cfg)
	if err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}

	select {}
}
