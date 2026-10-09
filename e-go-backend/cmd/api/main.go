package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/md-siam-kazi/e-go-backend/internal/config"
	"github.com/md-siam-kazi/e-go-backend/internal/database"
	"github.com/md-siam-kazi/e-go-backend/internal/server"
	"github.com/md-siam-kazi/e-go-backend/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	// Load .env for local development; ignore when absent (production uses
	// real environment variables).
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg.Log.Level)
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := database.NewPostgres(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer db.Close()
	log.Info("connected to postgres")

	rdb, err := database.NewRedis(ctx, cfg.Redis)
	if err != nil {
		return err
	}
	if rdb != nil {
		defer func() { _ = rdb.Close() }()
		log.Info("connected to redis")
	} else {
		log.Warn("redis disabled: REDIS_URL not set")
	}

	srv := server.New(cfg, log, db, rdb)
	httpServer := srv.HTTPServer()

	errCh := make(chan error, 1)
	go func() {
		log.Info("http server listening", slog.String("addr", httpServer.Addr), slog.String("env", cfg.App.Env))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return err
	}

	log.Info("server stopped cleanly")
	return nil
}
