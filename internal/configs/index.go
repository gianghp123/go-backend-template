package configs

import (
	"log"
	"os"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Logger   LoggerConfig
}

func Load() *Config {
	// Optional: load .env file with a library like godotenv.
	// Without it, rely on system environment variables.
	if _, err := os.Stat(".env"); err == nil {
		log.Println(".env file found, reading env vars from system")
	}

	return &Config{
		Server:   loadServerConfig(),
		Database: loadDatabaseConfig(),
		Auth:     loadAuthConfig(),
		Logger:   loadLoggerConfig(),
	}
}
