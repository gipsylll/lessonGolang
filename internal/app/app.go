package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sushkov/internal/config"
	"sushkov/internal/infrastructure"
	"sushkov/internal/logger"
	"sushkov/migrations"

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

type App struct {
	db     *pgxpool.Pool
	redis  *redis.Client
	server *http.Server
}

func New(cfg *config.Config) (*App, error) {
	if err := logger.Init(logger.Config{
		Level:   cfg.Logger.Level,
		Pretty:  cfg.Logger.Pretty,
		Service: "sushkov-api",
	}); err != nil {
		return nil, err
	}

	db, err := infrastructure.NewPostgresPool(&cfg.DB)
	if err != nil {
		return nil, err
	}

	rdb, err := infrastructure.NewRedisClient(&cfg.Redis)
	if err != nil {
		db.Close()
		return nil, err
	}

	if err := infrastructure.RunMigrations(db, migrations.FS); err != nil {
		db.Close()
		_ = rdb.Close()
		return nil, err
	}

	mux := http.NewServeMux()
	initUserHandler(db).Register(mux)

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      logger.LoggingMiddleware(mux),
		ReadTimeout:  _readTimeout,
		WriteTimeout: _writeTimeout,
		IdleTimeout:  _idleTimeout,
	}

	return &App{db: db, redis: rdb, server: srv}, nil
}

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
