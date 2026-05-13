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
	"github.com/joajo13/go-tracker/internal/broadcaster"
	"github.com/joajo13/go-tracker/internal/clock"
	"github.com/joajo13/go-tracker/internal/config"
	"github.com/joajo13/go-tracker/internal/domain"
	"github.com/joajo13/go-tracker/internal/persistence/sqlite"
	"github.com/joajo13/go-tracker/internal/scheduler"
	"github.com/joajo13/go-tracker/internal/sources"
	"github.com/joajo13/go-tracker/internal/web"
	"github.com/joajo13/go-tracker/internal/workers"
)

const shutdownTimeout = 5 * time.Second

func main() {
	if err := run(); err != nil {
		slog.Default().Error("agent exited with error", "err", err)
		os.Exit(1)
	}
}

//nolint:cyclop // wiring entrypoint; linear flow is intentional.
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger := newLogger(&cfg)
	slog.SetDefault(logger)

	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	db, err := sqlite.Open(rootCtx, cfg.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	tickerRepo := sqlite.NewTickerRepo(db)
	priceRepo := sqlite.NewPriceRepo(db)

	hub := broadcaster.New()
	persistSub := hub.Subscribe("persist", 256)
	go persistEvents(rootCtx, priceRepo, persistSub, logger)

	srcMap := map[string]sources.PriceSource{
		"yahoo": sources.NewYahoo(sources.YahooConfig{
			RatePerSec: cfg.Yahoo.RatePerSec, RateBurst: cfg.Yahoo.RateBurst,
		}),
		"dolarapi": sources.NewDolarAPI(sources.DolarAPIConfig{
			RatePerSec: cfg.DolarAPI.RatePerSec, RateBurst: cfg.DolarAPI.RateBurst,
		}),
	}

	jobs := make(chan sources.PollJob, 256)
	pool := workers.NewPool(workers.PoolConfig{
		Size:        cfg.WorkerPoolSize,
		Broadcaster: hub,
		Sources:     srcMap,
		JobTimeout:  10 * time.Second,
	})
	go func() {
		if poolErr := pool.Run(rootCtx, jobs); poolErr != nil && !errors.Is(poolErr, context.Canceled) {
			logger.Error("worker pool exited with error", "err", poolErr)
		}
	}()

	sched := scheduler.New(scheduler.Config{Clock: clock.Real{}, Jobs: jobs})
	active, err := tickerRepo.ListActive(rootCtx)
	if err != nil {
		return err
	}
	for i := range active {
		sched.Add(&active[i])
	}
	go func() {
		if schedErr := sched.Run(rootCtx); schedErr != nil && !errors.Is(schedErr, context.Canceled) {
			logger.Error("scheduler exited with error", "err", schedErr)
		}
	}()

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           newRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", cfg.HTTPAddr)
		if srvErr := srv.ListenAndServe(); srvErr != nil && !errors.Is(srvErr, http.ErrServerClosed) {
			errCh <- srvErr
		}
		close(errCh)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-stop:
		logger.Info("shutdown signal received")
	case srvErr := <-errCh:
		rootCancel()
		return srvErr
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	shutdownErr := srv.Shutdown(shutdownCtx)
	cancel()
	rootCancel()
	if shutdownErr != nil {
		return shutdownErr
	}
	logger.Info("server stopped cleanly")
	return nil
}

func persistEvents(ctx context.Context, repo *sqlite.PriceRepo, in <-chan domain.PriceEvent, logger *slog.Logger) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-in:
			if !ok {
				return
			}
			price := ev.Price
			if insertErr := repo.Insert(ctx, &price); insertErr != nil {
				logger.Warn("price_insert_failed",
					"ticker_id", ev.Price.TickerID,
					"source", ev.Price.Source,
					"err", insertErr)
			}
		}
	}
}

func newLogger(cfg *config.Config) *slog.Logger {
	level := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		level = slog.LevelDebug
	}
	opts := &slog.HandlerOptions{Level: level}
	if cfg.LogFormat == "text" {
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
