package utils

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func NewPostgresPool(ctx context.Context) (*pgxpool.Pool, error) {
	if err := godotenv.Load("config.env"); err != nil {
		log.Printf("Warning: .env file not loaded: %v. Make sure environment variables are set externally.", err)
	}

	var missing []string

	host := os.Getenv("DB_HOST")
	if host == "" {
		missing = append(missing, "DB_HOST")
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		missing = append(missing, "DB_PORT")
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		missing = append(missing, "DB_USER")
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		missing = append(missing, "DB_PASSWORD")
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		missing = append(missing, "DB_NAME")
	}

	if len(missing) > 0 {
		log.Printf("Error: Missing required environment variables: %v", missing)
		log.Print("Please set them in config.env or directly in the environment.")
		return nil, fmt.Errorf("missing required DB environment variables: %v", missing)
	}

	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbname,
	)

	log.Println("Connecting to database with DSN (password hidden):",
		fmt.Sprintf("postgresql://%s:***@%s:%s/%s?sslmode=disable", user, host, port, dbname))

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse DB config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create DB pool: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to connect DB pool: %w", err)
	}

	return pool, nil
}
