package config

import (
	"strconv"

	platformTypes "github.com/go-web-services/go-web-platform/types"
	platformUtils "github.com/go-web-services/go-web-platform/utils"
)

type AppConfig struct {
	Port            int
	Env             platformTypes.Environment
	MinioEndpoint   string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioUseSSL     bool
	SwaggerBasePath string
	TestEnv         string
}

type Config struct {
	App AppConfig
}

var Cfg Config

func LoadConfig() (*Config, error) {
	// Load from env variables or fallback to defaults
	portStr := platformUtils.GetEnv("APP_PORT", "8080")
	port, _ := strconv.Atoi(portStr)

	Cfg = Config{
		App: AppConfig{
			Port:            port,
			Env:             platformTypes.Environment(platformUtils.GetEnv("APP_ENV", "development")),
			MinioEndpoint:   platformUtils.GetEnv("MINIO_ENDPOINT", "localhost:9000"),
			MinioAccessKey:  platformUtils.GetEnv("MINIO_ACCESS_KEY", "minioadmin"),
			MinioSecretKey:  platformUtils.GetEnv("MINIO_SECRET_KEY", "minioadmin"),
			MinioUseSSL:     platformUtils.GetEnv("MINIO_USE_SSL", "false") == "true",
			SwaggerBasePath: platformUtils.GetEnv("SWAGGER_BASE_PATH", ""),
			TestEnv:         platformUtils.GetEnv("TEST_ENV", ""),
		},
	}

	return &Cfg, nil
}
