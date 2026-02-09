package repository

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func SetUpPostgresConn() *pgxpool.Pool {
	// Connection URL, loaded from environment variables
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))

	// Create a context with a timeout for the connection attempt
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a connection pool with default configurations
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Printf("Unable to create connection pool: %v\n", err)
	}

	// Verify the connection is working
	err = pool.Ping(ctx)
	if err != nil {
		log.Printf("Unable to ping database: %v\n", err)
	}

	fmt.Println("Successfully connected to the database!")

	return pool
}
