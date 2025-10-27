package config

type Config struct {
	DBHost         string `env:"DB_HOST" envDefault:"localhost"`
	DBPort         string `env:"DB_PORT" envDefault:"5432"`
	DBUser         string `env:"DB_USER,required"`
	DBPassword     string `env:"DB_PASSWORD,required"`
	DBName         string `env:"DB_NAME,required"`
	ServerPort     string `env:"SERVER_PORT" envDefault:"8080"`
	MigrationsPath string `env:"MIGRATIONS_PATH" envDefault:"migrations"`
}
