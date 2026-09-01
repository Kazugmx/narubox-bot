package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3"
	auth "github.com/kazugmx/narubox-bot/internal/auth"
	bot "github.com/kazugmx/narubox-bot/internal/bot"
	database "github.com/kazugmx/narubox-bot/internal/db"
)

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

	//database init
	pool, err := database.InitPool(ctx)
	if err != nil {
		return fmt.Errorf("initialize database :%w", err)
	}
	defer pool.Close()

	app := fiber.New()
	apiEndpoint := app.Group("/api/v1")

	//initialize routes
	auth.Route(apiEndpoint, pool)
	bot.Route(apiEndpoint, pool)

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

func initLogger() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	slog.SetDefault(logger)
	return logger
}
