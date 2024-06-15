package main

import (
	"hime-backend/db"
	"hime-backend/handler"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func init() {
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
	log.Fatal(app.Listen(":8080"))

	defer db.ClosePG()
}
