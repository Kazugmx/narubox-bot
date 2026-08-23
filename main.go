package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/Kazugmx/narubox-bot/db"
	Auth "github.com/Kazugmx/narubox-bot/svc/auth"
	"github.com/Kazugmx/narubox-bot/svc/auth/jwt"
	"github.com/Kazugmx/narubox-bot/svc/bot"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	app := fiber.New()
	ctx := context.Background()
	jwtService, err := jwt.NewJWTService(
		os.Getenv("JWT_SECRET"),
		os.Getenv("CALLBACK_ORIGIN"),
	)
	if err != nil {
		log.Fatal(err)
	}

	query, conn, err := initiateDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	apiRoute := app.Group("/api/v1")

	Auth.Route(apiRoute, query, jwtService)
	botService := bot.NewService(query, http.DefaultClient, os.Getenv("YOUTUBE_API_KEY"), os.Getenv("CALLBACK_ORIGIN"), slog.Default())
	bot.Route(apiRoute, botService, jwtService)

	defer conn.Close()
	log.Fatal(app.Listen(":3000"))
}

// Initiate conn with Database
func buildDBURL() (string, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")

	if !(len(host) > 0) || !(len(port) > 0) || !(len(user) > 0) || !(len(pass) > 0) || !(len(name) > 0) {
		return "", fmt.Errorf("database environment variables are not set")
	}

	dbURL := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pass),
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   name,
	}
	if _, err := url.ParseRequestURI(dbURL.String()); err != nil {
		return "", fmt.Errorf("invalid database URL: %w", err)
	}
	return dbURL.String(), nil
}

func initiateDatabase(ctx context.Context) (*db.Queries, *pgxpool.Pool, error) {
	dbURL, err := buildDBURL()
	if err != nil {
		return nil, nil, err
	}
	conn, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, nil, fmt.Errorf("create database pool: %w", err)
	}
	if err := conn.Ping(ctx); err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("connect to database: %w", err)
	}

	return db.New(conn), conn, nil
}
