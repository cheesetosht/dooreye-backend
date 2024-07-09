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
	envVars := []string{"PORT", "DATABASE_URL", "JWT_SECRET", "AWS_REGION", "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_S3_BUCKET_NAME"}
	utility.ValidateEnv(envVars)
	log.Println("> start program")
}

func main() {
	// database connection
	db.InitPG()

	_ = utility.GetFirebaseApp()

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
	app.Post("/visitors/from-residence", handler.CreateVisitorFromResidence, middleware.AuthByRoleLevel(1))

	app.Get("/visitors", handler.GetVisitor, middleware.AuthByRoleLevel(2))
	app.Post("/visits", handler.CreateVisitByVisitorID, middleware.AuthByRoleLevel(2))
	app.Post("/visits/new", handler.CreateVisitForNewVisitor, middleware.AuthByRoleLevel(2))

	app.Post("/users/resident-bulk", handler.BulkCreateResidentUser, middleware.AuthByRoleLevel(4))
	// app.Post("/cities", handler.InsertCity)
	app.Post("/societies", handler.CreateSociety, middleware.AuthByRoleLevel(5))
	app.Get("/cities", handler.FetchCities, middleware.AuthByRoleLevel(5))
	app.Get("/societies", handler.FetchSocieties, middleware.AuthByRoleLevel(5))
	app.Post("/blocks/bulk", handler.BulkCreateBlocks, middleware.AuthByRoleLevel(5))
	app.Post("/residences/bulk", handler.BulkCreateResidences, middleware.AuthByRoleLevel(5))

	app.Post("/push-notifications", handler.SendPushNotification, middleware.AuthByRoleLevel(5))

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
