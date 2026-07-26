package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	neturl "net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prayogopangestu/crm-system/backend/configs"
	redisinfra "github.com/prayogopangestu/crm-system/backend/internal/infrastructure/cache/redis"
	dbpostgres "github.com/prayogopangestu/crm-system/backend/internal/infrastructure/database/postgres"
	"github.com/prayogopangestu/crm-system/backend/internal/infrastructure/jwt"
	loggerinfra "github.com/prayogopangestu/crm-system/backend/internal/infrastructure/logger"
	"github.com/prayogopangestu/crm-system/backend/internal/infrastructure/telegram"
	pgrepo "github.com/prayogopangestu/crm-system/backend/internal/repository/postgres"
	httproutes "github.com/prayogopangestu/crm-system/backend/internal/delivery/http/routes"
	integrationusecase "github.com/prayogopangestu/crm-system/backend/internal/usecase/integration"
	usecasesupport "github.com/prayogopangestu/crm-system/backend/internal/usecase/support"
	"github.com/prayogopangestu/crm-system/backend/pkg/encryption"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type application struct {
	cfg        configs.Config
	log        *slog.Logger
	db         *gorm.DB
	cache      *redisinfra.Cache
	httpServer *http.Server
	worker     *integrationusecase.Worker
}

func main() {
	cfg, configPath, err := loadConfig()
	if err != nil {
		slog.Error("load config failed", "error", err)
		os.Exit(1)
	}

	log := loggerinfra.New(cfg.App.LogLevel)
	slog.SetDefault(log)
	logConfig(log, cfg, configPath)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := newApplication(ctx, cfg, log)
	if err != nil {
		log.Error("application startup failed", "error", err)
		os.Exit(1)
	}
	defer app.close()

	app.run(ctx, stop)
}

func loadConfig() (configs.Config, string, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}
	cfg, err := configs.Load(configPath)
	return cfg, configPath, err
}

func logConfig(log *slog.Logger, cfg configs.Config, configPath string) {
	log.Info("database config resolved",
		"DATABASE_URL_set", os.Getenv("DATABASE_URL") != "",
		"DATABASE_URL_len", len(os.Getenv("DATABASE_URL")),
		"resolved_db_url", maskURL(cfg.Database.URL),
		"config_path", configPath,
	)
}

func newApplication(ctx context.Context, cfg configs.Config, log *slog.Logger) (*application, error) {
	location, err := time.LoadLocation(cfg.App.Timezone)
	if err != nil {
		return nil, err
	}

	db := initDB(cfg, log, location)
	cache := initCache(ctx, cfg, log)
	tokens := jwt.New(cfg.Auth.JWTSecret, cfg.Auth.JWTTTL)
	cipher, err := encryption.New(cfg.Security.EncryptionKey)
	if err != nil {
		if sqlDB, dbErr := db.DB(); dbErr == nil {
			_ = sqlDB.Close()
		}
		return nil, err
	}

	telegramClient := telegram.New()
	cacheHelper := usecasesupport.CacheHelper{Cache: cache, Logger: log}
	worker := initWorker(cfg, db, location, cipher, telegramClient, log)

	httpServer := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           httproutes.NewRouter(httproutes.Deps{
			DB:          db,
			Cache:       cache,
			CacheHelper: cacheHelper,
			Tokens:      tokens,
			Logger:      log,
			Location:    location,
			BaseURL:     cfg.App.BaseURL,
			BcryptCost:  cfg.Auth.BcryptCost,
			Origins:     cfg.HTTP.AllowedOrigins,
			Cipher:      cipher,
			Sender:      telegramClient,
			Ready:       func() error { sqlDB, err := db.DB(); if err != nil { return err }; return sqlDB.PingContext(context.Background()) },
		}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &application{
		cfg:        cfg,
		log:        log,
		db:         db,
		cache:      cache,
		httpServer: httpServer,
		worker:     worker,
	}, nil
}

func initDB(cfg configs.Config, log *slog.Logger, location *time.Location) *gorm.DB {
	sqlDB, err := dbpostgres.NewConnection(dbpostgres.Config{
		URL:          cfg.Database.URL,
		MaxOpenConns: int(cfg.Database.MaxConns),
		MaxIdleConns: int(cfg.Database.MinConns),
	})
	if err != nil {
		log.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
		TranslateError:         true,
	})
	if err != nil {
		log.Error("gorm open failed", "error", err)
		os.Exit(1)
	}
	return db
}

func initCache(ctx context.Context, cfg configs.Config, log *slog.Logger) *redisinfra.Cache {
	cache, err := redisinfra.New(cfg.Redis.URL)
	if err != nil {
		log.Warn("redis configuration invalid; cache disabled", "error", err)
		return nil
	}
	if err := cache.Ping(ctx); err != nil {
		log.Warn("redis unavailable; cache will fail open", "error", err)
	}
	return cache
}

func initWorker(
	cfg configs.Config,
	db *gorm.DB,
	location *time.Location,
	cipher *encryption.Cipher,
	telegramClient *telegram.Client,
	log *slog.Logger,
) *integrationusecase.Worker {
	integrationRepository := pgrepo.NewIntegrationRepository(db, location)
	return integrationusecase.NewWorker(
		integrationRepository, cipher, telegramClient, log,
		cfg.Telegram.WorkerInterval, cfg.Telegram.WorkerBatchSize,
	)
}

func (a *application) run(ctx context.Context, stop context.CancelFunc) {
	go a.worker.Run(ctx)
	go a.runHTTP(stop)

	<-ctx.Done()
	a.shutdown()
}

func (a *application) runHTTP(stop context.CancelFunc) {
	a.log.Info("binding HTTP server", "addr", a.cfg.HTTP.Addr, "PORT_env", os.Getenv("PORT"))
	a.log.Info("HTTP server started", "addr", a.cfg.HTTP.Addr)
	if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		a.log.Error("HTTP server failed", "error", err)
		stop()
	}
}

func (a *application) shutdown() {
	a.log.Info("shutdown started")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = a.httpServer.Shutdown(shutdownCtx)
	a.log.Info("shutdown complete")
}

func (a *application) close() {
	if a.cache != nil {
		_ = a.cache.Close()
	}
	if a.db != nil {
		if sqlDB, err := a.db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
}

func maskURL(rawURL string) string {
	u, err := neturl.Parse(rawURL)
	if err != nil {
		return "[unparseable]"
	}
	host := u.Host
	if host == "" {
		host = "[no-host]"
	}
	if u.User != nil {
		u.User = neturl.User(u.User.Username())
	}
	return "scheme=" + u.Scheme + " host=" + host
}
