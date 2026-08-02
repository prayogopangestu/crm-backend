package configs

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App struct {
		Env      string `yaml:"env"`
		BaseURL  string `yaml:"base_url"`
		Timezone string `yaml:"timezone"`
		LogLevel string `yaml:"log_level"`
	} `yaml:"app"`
	HTTP struct {
		Addr           string   `yaml:"addr"`
		AllowedOrigins []string `yaml:"allowed_origins"`
	} `yaml:"http"`
	Database struct {
		URL      string `yaml:"url"`
		MaxConns int32  `yaml:"max_conns"`
		MinConns int32  `yaml:"min_conns"`
	} `yaml:"database"`
	Redis struct {
		URL string `yaml:"url"`
	} `yaml:"redis"`
	Auth struct {
		JWTSecret  string        `yaml:"jwt_secret"`
		JWTTTL     time.Duration `yaml:"-"`
		JWTTTLText string        `yaml:"jwt_ttl"`
		BcryptCost int           `yaml:"bcrypt_cost"`
	} `yaml:"auth"`
	Security struct {
		EncryptionKey string `yaml:"encryption_key"`
	} `yaml:"security"`
	Google struct {
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
		RedirectURL  string `yaml:"redirect_url"`
	} `yaml:"google"`
	Telegram struct {
		WorkerInterval     time.Duration `yaml:"-"`
		WorkerIntervalText string        `yaml:"worker_interval"`
		WorkerBatchSize    int           `yaml:"worker_batch_size"`
	} `yaml:"telegram"`
}

func Load(path string) (Config, error) {
	var cfg Config
	raw, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return cfg, fmt.Errorf("decode config: %w", err)
	}

	override(&cfg)
	if cfg.Auth.JWTTTL, err = time.ParseDuration(cfg.Auth.JWTTTLText); err != nil {
		return cfg, fmt.Errorf("invalid jwt ttl: %w", err)
	}
	if cfg.Telegram.WorkerInterval, err = time.ParseDuration(cfg.Telegram.WorkerIntervalText); err != nil {
		return cfg, fmt.Errorf("invalid telegram worker interval: %w", err)
	}
	if len(cfg.Auth.JWTSecret) < 32 {
		return cfg, errors.New("JWT_SECRET must be at least 32 characters")
	}
	key, err := base64.StdEncoding.DecodeString(cfg.Security.EncryptionKey)
	if err != nil || len(key) != 32 {
		return cfg, errors.New("APP_ENCRYPTION_KEY must be base64 encoded 32 bytes")
	}
	if _, err := time.LoadLocation(cfg.App.Timezone); err != nil {
		return cfg, fmt.Errorf("invalid timezone: %w", err)
	}
	return cfg, nil
}
