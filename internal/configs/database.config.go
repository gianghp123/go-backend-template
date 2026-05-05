package configs

import "your-project/internal/utils"

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     utils.GetEnv("DB_HOST", "localhost"),
		Port:     utils.GetEnvInt("DB_PORT", 5432),
		User:     utils.GetEnv("DB_USER", "postgres"),
		Password: utils.GetEnv("DB_PASSWORD", "postgres"),
		DBName:   utils.GetEnv("DB_NAME", "myapp"),
		SSLMode:  utils.GetEnv("DB_SSLMODE", "disable"),
	}
}
