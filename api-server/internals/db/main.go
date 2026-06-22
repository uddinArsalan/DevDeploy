package db

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDB(ctx context.Context) *pgxpool.Pool {
	connStr := os.Getenv("DB_CONN_STR")
	if connStr == "" {
		log.Fatalf("connection string not set")
	}
	dbInstance, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("Critical error establishing database connection: %v", err)
	}
	if err = dbInstance.Ping(ctx); err != nil {
		log.Fatalf("Database was unreachable on initialization: %v", err)
	}
	return dbInstance
}
