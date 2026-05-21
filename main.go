package main

import (
	"context"
	"fmt"
	"log"
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

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		user,
		pass,
		host,
		port,
		name,
	)
}

func initiateDatabase(ctx context.Context) (*db.Queries, *pgxpool.Pool) {

	dbUrl := buildDBUrl()
	if dbUrl == "" {
		log.Fatalln("[ERROR] env:DATABASE_URL is empty.")
	}
	conn, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		fmt.Printf("Unable to connect to database:%v\n", err)
		os.Exit(1)
	}

	query := db.New(conn)

	return query, conn
}
