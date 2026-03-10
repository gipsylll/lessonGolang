package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sushkov/internal/adapter/memory"
	"sushkov/internal/config"
	"sushkov/internal/domain"
	"sushkov/internal/handlers"
	"sushkov/internal/infrastructure"
	"sushkov/internal/logger"
	"sushkov/internal/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	_shutdownTimeout = 10 * time.Second
	_readTimeout     = 5 * time.Second
	_writeTimeout    = 10 * time.Second
	_idleTimeout     = 60 * time.Second
)

// App — центральная структура приложения, владеет всеми зависимостями.
type App struct {
	db     *pgxpool.Pool
	redis  *redis.Client
	server *http.Server
}

// New инициализирует зависимости через DI и возвращает готовое приложение.
func New(cfg *config.Config) (*App, error) {
	// 1. Логгер
	if err := logger.Init(logger.Config{
		Level:   cfg.Logger.Level,
		Pretty:  cfg.Logger.Pretty,
		Service: "sushkov-api",
	}); err != nil {
		return nil, err
	}

	// 2. Инфраструктура
	db, err := infrastructure.NewPostgresPool(&cfg.DB)
	if err != nil {
		return nil, err
	}

	rdb, err := infrastructure.NewRedisClient(&cfg.Redis)
	if err != nil {
		db.Close()
		return nil, err
	}

	// 3. TODO: Миграции
	// if err := infrastructure.RunMigrations(db); err != nil { ... }

	// 4. DI: repo → usecase → handler
	userRepo := memory.NewUserRepo([]domain.User{ // TODO: заменить на postgres.NewUserRepo(db)
		{ID: 1, Name: "John Doe", Email: "john.doe@example.com"},
		{ID: 2, Name: "Jane Doe", Email: "jane.doe@example.com"},
	})
	userUsecase := usecase.NewUserUsecase(userRepo)
	userHandler := handlers.NewUserHandler(userUsecase)

	// 5. Роутер — каждый хендлер сам регистрирует свои роуты
	mux := http.NewServeMux()
	userHandler.Register(mux)

	// 6. HTTP-сервер с таймаутами
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      logger.LoggingMiddleware(mux),
		ReadTimeout:  _readTimeout,
		WriteTimeout: _writeTimeout,
		IdleTimeout:  _idleTimeout,
	}

	return &App{db: db, redis: rdb, server: srv}, nil
}

// Run запускает сервер и блокируется до SIGINT/SIGTERM, затем graceful shutdown.
func (a *App) Run() {
	go func() {
		log.Info().Str("addr", a.server.Addr).Msg("server starting")
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Info().Str("signal", sig.String()).Msg("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), _shutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server shutdown error")
	}

	a.db.Close()
	if err := a.redis.Close(); err != nil {
		log.Error().Err(err).Msg("redis close error")
	}

	log.Info().Msg("server stopped")
}
