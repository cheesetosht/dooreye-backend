package main

import (
	"hime-backend/db"
	"hime-backend/handler"
	"hime-backend/middleware"
	"hime-backend/utility"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("!! error loading .env file")
	}
	envVars := []string{"PORT", "DATABASE_URL", "JWT_SECRET"}
	utility.ValidateEnv(envVars)
	log.Println("> start program")
}

func main() {
	// database connection
	db.InitPG()

	// router
	app := fiber.New()
	app.Use(logger.New())
	app.Use(limiter.New(limiter.Config{
		Max:        48,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "rate limit exceeded, please try again later.",
			})
		},
	}))
	app.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("pong")
	})
	app.Post("/auth/request-otp", handler.RequestOTP)
	app.Post("/auth/verify-otp", handler.VerifyOTP)
	app.Post("/societies", handler.InsertSociety, middleware.AuthByRoleLevel(4))

	// app.Post("/cities", handler.InsertCity)
	app.Get("/cities", handler.GetCities, middleware.AuthByRoleLevel(5))
	app.Get("/societies", handler.GetSocieties, middleware.AuthByRoleLevel(5))
	app.Post("/blocks/bulk", handler.BulkInsertBlocks, middleware.AuthByRoleLevel(5))
	app.Post("/residences/bulk", handler.BulkInsertResidences, middleware.AuthByRoleLevel(5))
	app.Post("/users/resident-bulk", handler.BulkInsertResidentUser, middleware.AuthByRoleLevel(4))

	go func() {
		if err := app.Listen(":" + os.Getenv("PORT")); err != nil {
			log.Println("!! failed to listen to provided port: ", err)
			os.Exit(255)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	log.Println("> shutting down server...", <-quit)

	if err := app.Shutdown(); err != nil {
		log.Fatal("> server forced to shutdown:\n", err)
	}

	log.Println("> server exiting")

	defer db.ClosePG()
}
