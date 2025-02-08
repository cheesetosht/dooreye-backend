package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"dooreye-backend/internal/store"
	"dooreye-backend/internal/api"
)

func main() {
	if err := run(); err != nil {
		slog.Error("startup error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// Load environment
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development"
	}

	envFile := ".env"
	if env != "production" {
		envFile = fmt.Sprintf(".env.%s", env)
	}
	if err := godotenv.Load(envFile); err != nil {
		fmt.Printf("No %s file found\n", envFile)
	}

	// Setup logger
	var logHandler slog.Handler
	if env == "development" {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	} else {
		logHandler = slog.NewJSONHandler(os.Stdout, nil)
	}
	log := slog.New(logHandler)

	// Initialize store with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := store.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	server := api.NewHandler(db, log)

	serverErrors := make(chan error, 1)
	go func() {
		log.Info("starting server",
			"port", os.Getenv("PORT"),
			"env", env,
		)
		serverErrors <- server.Run(":" + os.Getenv("PORT"))
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Info("shutdown signal received", "signal", sig)

		// Give outstanding requests a deadline for completion
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Shutdown gracefully
		if err := server.Shutdown(ctx); err != nil {
			server.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
