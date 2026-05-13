// Package main is the portfolio-agent entrypoint.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/joajo13/go-tracker/internal/api"
	"github.com/joajo13/go-tracker/internal/web"
)

const shutdownTimeout = 5 * time.Second

func main() {
	logger := newLogger()
	slog.SetDefault(logger)

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           newRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-stop:
		logger.Info("shutdown signal received")
	case err := <-errCh:
		logger.Error("server error", "err", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	shutdownErr := srv.Shutdown(ctx)
	cancel()
	if shutdownErr != nil {
		logger.Error("graceful shutdown failed", "err", shutdownErr)
		os.Exit(1)
	}
	logger.Info("server stopped cleanly")
}

func newLogger() *slog.Logger {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{Level: level}
	if os.Getenv("LOG_FORMAT") == "text" {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}

func newRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Handle("/healthz", api.HealthHandler())

	if h, err := web.Handler(); err == nil {
		r.Mount("/", h)
	} else {
		slog.Default().Error("failed to mount embedded frontend", "err", err)
	}

	return r
}
