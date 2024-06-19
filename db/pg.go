package db

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	PGPool *pgxpool.Pool
	once   sync.Once
)

func config() *pgxpool.Config {
	const defaultMaxConns = int32(4)
	const defaultMinConns = int32(0)
	const defaultMaxConnLifetime = time.Hour
	const defaultMaxConnIdleTime = time.Minute * 30
	const defaultHealthCheckPeriod = time.Minute
	const defaultConnectTimeout = time.Second * 5

	DATABASE_URL := os.Getenv("DATABASE_URL")

	dbConfig, err := pgxpool.ParseConfig(DATABASE_URL)
	if err != nil {
		log.Fatal("!! failed to create pgx config, error:\n", err)
	}

	dbConfig.MaxConns = defaultMaxConns
	dbConfig.MinConns = defaultMinConns
	dbConfig.MaxConnLifetime = defaultMaxConnLifetime
	dbConfig.MaxConnIdleTime = defaultMaxConnIdleTime
	dbConfig.HealthCheckPeriod = defaultHealthCheckPeriod
	dbConfig.ConnConfig.ConnectTimeout = defaultConnectTimeout

	dbConfig.BeforeAcquire = func(ctx context.Context, c *pgx.Conn) bool {
		log.Println("> before conn. pool acquire")
		return true
	}

	dbConfig.AfterRelease = func(c *pgx.Conn) bool {
		log.Println("> after conn. pool acquire")
		return true
	}

	dbConfig.BeforeClose = func(c *pgx.Conn) {
		log.Println("> closed db conn. pool")
	}

	return dbConfig
}

func InitPG() {
	once.Do(func() {
		var err error
		PGPool, err = pgxpool.NewWithConfig(context.Background(), config())
		if err != nil {
			log.Fatalf("!! unable to create connection pool: %v\n", err)
		}

		err = PGPool.Ping(context.Background())
		if err != nil {
			log.Fatalf("!! unable to connect to database: %v\n", err)
		}
		log.Println("> database connected")
	})
}

func ClosePG() {
	if PGPool != nil {
		PGPool.Close()
	}
}
