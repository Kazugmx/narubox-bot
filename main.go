package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/Kazugmx/narubox-bot/db"
	jwtOperator "github.com/Kazugmx/narubox-bot/internal/auth"
	Auth "github.com/Kazugmx/narubox-bot/svc/auth"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	app := fiber.New()
	ctx := context.Background()
	jwtService := jwtOperator.NewJWTService(os.Getenv("JWT_SECRET"))

	query, conn := initiateDatabase(ctx)
	apiRoute := app.Group("/api/v1")

	Auth.Route(apiRoute, query, ctx, jwtService)

	defer conn.Close()
	log.Fatal(app.Listen(":3000"))
}

// Initiate conn with Database
func buildDBUrl() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")

	if !(len(host) > 0) || !(len(port) > 0) || !(len(user) > 0) || !(len(pass) > 0) || !(len(name) > 0) {
		log.Fatalln("error\t database environment variables are not set.")
	}

	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		user,
		pass,
		host,
		port,
		name,
	)
	_, err := url.ParseRequestURI(dbUrl)
	if err != nil {
		log.Fatalf("invalid dbUrl: %v\n", err)
	}
	return dbUrl
}

func initiateDatabase(ctx context.Context) (*db.Queries, *pgxpool.Pool) {

	dbUrl := buildDBUrl()
	if dbUrl == "" {
		log.Fatalln("[ERROR] env:DATABASE_URL is empty.")
	}
	conn, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	query := db.New(conn)

	return query, conn
}
