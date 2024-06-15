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
	app.Get("/cities", handler.GetCities)
	// app.Post("/cities", handler.InsertCity)
	app.Post("/societies", handler.InsertSociety)
	app.Get("/societies", handler.GetSocieties)
	app.Post("/blocks/bulk", handler.BulkInsertBlocks)
	app.Post("/residences/bulk", handler.BulkInsertResidences)

	log.Fatal(app.Listen(":8080"))

	defer db.ClosePG()
}
