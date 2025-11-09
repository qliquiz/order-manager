package main

import (
	"349877-artemkagor05-course-1478/internal/app"
	"349877-artemkagor05-course-1478/internal/cache"
	"349877-artemkagor05-course-1478/internal/config"
	"349877-artemkagor05-course-1478/internal/lib/logger/handlers/slogpretty"
	"349877-artemkagor05-course-1478/internal/postgres"
	"349877-artemkagor05-course-1478/internal/redis"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	envLocal = "local"
	envProd  = "prod"
	envDev   = "dev"
)

func main() {
	cfg := config.MustLoad()

	initCtx := context.Background()
	redisClient, err := redis.New(initCtx, cfg.Redis)
	if err != nil {
		fmt.Printf("failed to initialize redis client: %e", err)
		os.Exit(1)
	}
	defer func() {
		if err = redisClient.Close(); err != nil {
			fmt.Printf("failed to close redis connection: %e", err)
		}
	}()

	db, err := postgres.New(cfg.DB)
	if err != nil {
		panic(fmt.Errorf("failed to connect to the database: %w", err))
	}

	log := setupLogger(cfg.Env)
	log.Info("starting application...")

	application := app.New(db.Pool, cache.New(redisClient), log, cfg.GRPC.Port, cfg.Gateway.Port, cfg.GRPC.Timeout)
	go application.GrpcApp.MustRun()
	go application.GatewayApp.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	log.Info("shutting down application...", slog.String("signal", sign.String()))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err = application.GatewayApp.Stop(shutdownCtx); err != nil {
		log.Error("gateway shutdown error", "err", err)
	}
	if err = application.GrpcApp.Stop(shutdownCtx); err != nil {
		log.Error("application shutdown error", "err", err)
	}
	db.Close()
	log.Info("database connection pool closed")

	log.Info("Server Stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
