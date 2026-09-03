package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/kazugmx/narubox-bot/internal/auth"
	"github.com/kazugmx/narubox-bot/internal/bot"
	"github.com/kazugmx/narubox-bot/internal/config"
	database "github.com/kazugmx/narubox-bot/internal/db"
)

type AppConfig = config.AppConfig

func main() {
	//initialize requirements
	logger := initLogger()

	if err := run(); err != nil {
		logger.Error("application terminated", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := defineContext()
	defer stop()

	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	//database init
	pool, err := database.InitPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("initialize database :%w", err)
	}
	defer pool.Close()

	if len(os.Args) > 1 && os.Args[1] == "init" {
		return config.RunMigration(ctx, *cfg, pool)
	}

	app := fiber.New()
	apiEndpoint := app.Group("/api/v1")

	//initialize routes
	auth.Route(apiEndpoint, pool)
	bot.Route(apiEndpoint, pool)

	go func() { //goroutineなんもわからん
		<-ctx.Done()
		slog.Info("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			slog.Error("failed to shutdown server", "error", err)
		}

	}()

	if err := app.Listen(":3000"); err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	return nil
}

func defineContext() (ctx context.Context, stop context.CancelFunc) {
	return signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
}

func LoadConfig() (*AppConfig, error) {
	cfg := &AppConfig{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		JWTIssuer:      os.Getenv("JWT_ISSUER"),
		CallbackOrigin: os.Getenv("CALLBACK_ORIGIN"),
		URIMaster:      os.Getenv("URI_MASTER"),
		APIKey:         os.Getenv("APIKEY"),
	}

	var errs []string

	if cfg.DatabaseURL == "" {
		errs = append(errs, "DATABASE_URL is not set")
	}

	if cfg.JWTSecret == "" {
		errs = append(errs, "JWT_SECRET is not set")
	}

	if cfg.JWTIssuer == "" {
		errs = append(errs, "JWT_ISSUER is not set")
	}
	if cfg.CallbackOrigin == "" {
		errs = append(errs, "CALLBACK_ORIGIN is not set")
	}
	if cfg.URIMaster == "" {
		errs = append(errs, "URI_MASTER is not set")
	}
	if cfg.APIKey == "" {
		errs = append(errs, "APIKEY is not set (Youtube Data API v3 is required.)")
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("invalid configuration: %s", strings.Join(errs, "; "))
	}

	return cfg, nil
}

func initLogger() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	slog.SetDefault(logger)
	return logger
}
