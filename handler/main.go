package handler

import (
	"context"
	"fmt"
	"hime-backend/db"
	"hime-backend/models"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

type cityPresenter struct {
	ID    int32  `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

func GetCities(c fiber.Ctx) error {
	searchStr := c.Query("search")

	if searchStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "search parameter is required",
		})
	}
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid page parameter",
		})
	}
	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid limit parameter",
		})
	}
	offset := (page - 1) * limit

	query := `SELECT id, name, state FROM cities c WHERE lower(c.name) LIKE $1 LIMIT $2 OFFSET $3;`
	rows, err := db.PGPool.Query(context.Background(), query, fmt.Sprintf("%s%%", searchStr), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	cities, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByPos[cityPresenter])
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	return c.JSON(&fiber.Map{
		"data":    cities,
		"error":   nil,
		"success": true,
	})
}

func InsertCity(c fiber.Ctx) error {
	// Parse request body into City struct
	var city models.City

	if err := c.Bind().Body(&city); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid input",
		})
	}

	if city.Name == "" || city.State == "" || city.Country == "" || city.Lat == 0 || city.Lng == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "all fields are required",
		})
	}
	fmt.Println(city.Name)

	query := `INSERT INTO cities (name, state, country, lat, lng) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := db.PGPool.QueryRow(context.Background(), query, city.Name, city.State, city.Country, city.Lat, city.Lng).Scan(&city.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not insert city",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    false,
	})
}
