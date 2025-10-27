package postgres

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/devopesik/wallet-basic-operations/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	goose "github.com/pressly/goose/v3"
)

func RunMigrations(cfg *config.Config) error {
	log.Println("Запуск миграций...")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("не удалось открыть подключение для миграций: %w", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Printf("ошибка при закрытии подключения к БД для миграций: %v", err)
		}
	}(db)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("не удалось подключиться к БД для миграций: %w", err)
	}

	err = goose.SetDialect("postgres")
	if err != nil {
		return err
	}
	if err := goose.Up(db, cfg.MigrationsPath); err != nil {
		return fmt.Errorf("ошибка при применении миграций: %w", err)
	}

	log.Println("Все миграции успешно применены")
	return nil
}
