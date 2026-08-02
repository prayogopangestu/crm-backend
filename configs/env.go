package configs

import (
	"os"
	"strconv"
	"strings"
)

func override(cfg *Config) {
	setString("APP_ENV", &cfg.App.Env)
	setString("APP_BASE_URL", &cfg.App.BaseURL)
	setString("APP_TIMEZONE", &cfg.App.Timezone)
	setString("LOG_LEVEL", &cfg.App.LogLevel)
	setString("HTTP_ADDR", &cfg.HTTP.Addr)
	if port := os.Getenv("PORT"); port != "" && os.Getenv("HTTP_ADDR") == "" {
		cfg.HTTP.Addr = ":" + port
	}
	if cfg.HTTP.Addr == "" {
		cfg.HTTP.Addr = ":8080"
	}
	setString("DATABASE_URL", &cfg.Database.URL)
	setString("REDIS_URL", &cfg.Redis.URL)
	setString("JWT_SECRET", &cfg.Auth.JWTSecret)
	setString("JWT_TTL", &cfg.Auth.JWTTTLText)
	setString("APP_ENCRYPTION_KEY", &cfg.Security.EncryptionKey)
	setString("GOOGLE_CLIENT_ID", &cfg.Google.ClientID)
	setString("GOOGLE_CLIENT_SECRET", &cfg.Google.ClientSecret)
	setString("GOOGLE_REDIRECT_URL", &cfg.Google.RedirectURL)
	setString("TELEGRAM_WORKER_INTERVAL", &cfg.Telegram.WorkerIntervalText)
	if v := os.Getenv("CORS_ALLOWED_ORIGINS"); v != "" {
		cfg.HTTP.AllowedOrigins = strings.Split(v, ",")
	}
	setInt32("DATABASE_MAX_CONNS", &cfg.Database.MaxConns)
	setInt32("DATABASE_MIN_CONNS", &cfg.Database.MinConns)
	setInt("BCRYPT_COST", &cfg.Auth.BcryptCost)
	setInt("TELEGRAM_WORKER_BATCH_SIZE", &cfg.Telegram.WorkerBatchSize)
}

func setString(key string, dst *string) {
	if value := os.Getenv(key); value != "" {
		*dst = value
	}
}

func setInt(key string, dst *int) {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			*dst = parsed
		}
	}
}

func setInt32(key string, dst *int32) {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 32); err == nil {
			*dst = int32(parsed)
		}
	}
}
