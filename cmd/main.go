package main

import (
	"hime-backend/db"
	"hime-backend/handler"
	"hime-backend/utility"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("!! error loading .env file")
	}
	envVars := []string{"PORT", "DATABASE_URL"}
	utility.ValidateEnv(envVars)
	log.Println("> start program")
}

func main() {
	// database connection
	db.InitPG()

	// router
	app := fiber.New()
	app.Use(logger.New())
	app.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("pong")
	})
	app.Get("/cities", handler.GetCities)
	// app.Post("/cities", handler.InsertCity)
	app.Post("/societies", handler.InsertSociety)
	app.Get("/societies", handler.GetSocieties)
	app.Post("/blocks/bulk", handler.BulkInsertBlocks)
	app.Post("/residences/bulk", handler.BulkInsertResidences)
	app.Post("/residents", handler.InsertResident)

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
