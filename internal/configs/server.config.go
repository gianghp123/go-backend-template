package configs

import "your-project/internal/utils"

type ServerConfig struct {
	Port string
	Mode string
}

func loadServerConfig() ServerConfig {
	return ServerConfig{
		Port: utils.GetEnv("PORT", "3001"),
		Mode: utils.GetEnv("GIN_MODE", "debug"),
	}
}
