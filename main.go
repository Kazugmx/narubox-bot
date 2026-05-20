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
	apiRoute := app.Group("/api/v1")
	ctx := context.Background()

	query, conn := initializeDatabase(ctx)

	jwtOperator.InitAuth()

	var greeting string
	err := conn.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)
	if err != nil {
		log.Fatalln("error :", err)
	}

	app.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"hello": greeting})
	})

	Auth.Route(apiRoute, query, ctx)

	defer conn.Close()
	log.Fatal(app.Listen(":3000"))
}

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

func initializeDatabase(ctx context.Context) (*db.Queries, *pgxpool.Pool) {

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
