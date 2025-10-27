package config

import (
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

func Load(path string) *Config {
	if err := godotenv.Load(path); err != nil {
		log.Printf("%s файл не найден, используются только системные переменные", path)
	}

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Ошибка при загрузке конфигурации: %+v", err)
	}

	log.Println("Конфигурация успешно загружена")
	return &cfg
}
