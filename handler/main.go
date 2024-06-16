package handler

import (
	"context"
	"fmt"
	"hime-backend/db"
	"hime-backend/models"
	"hime-backend/repository"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

func InsertCity(c fiber.Ctx) error {
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

type societyCollector struct {
	Name          string `json:"name"`
	Developer     string `json:"developer"`
	MaxResidences int32  `json:"max_residences"`
	CityID        int32  `json:"city_id"`
}

func InsertSociety(c fiber.Ctx) error {
	var society societyCollector

	if err := c.Bind().Body(&society); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid input",
		})
	}

	if society.Name == "" || society.Developer == "" || society.MaxResidences == 0 || society.CityID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "all fields are required",
		})
	}

	var id int
	query := `INSERT INTO societies (name, developer, max_residences, city_id) VALUES ($1, $2, $3, $4) RETURNING id`
	err := db.PGPool.QueryRow(context.Background(), query, society.Name, society.Developer, society.MaxResidences, society.CityID).Scan(&id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not create society",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"id": id,
		},
	})
}

type societyPresenter struct {
	ID            int32  `json:"id"`
	Name          string `json:"name"`
	Developer     string `json:"developer"`
	MaxResidences int32  `json:"max_residences"`
	City          string `json:"city"`
}

func GetSocieties(c fiber.Ctx) error {
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

	query := `SELECT s.id, s.name, developer, max_residences, c.name city
	FROM societies s
		left join cities c on s.city_id = c.id
	WHERE s.is_valid = true
	  	and lower(s.name) LIKE $1
	LIMIT $2 OFFSET $3;`
	rows, err := db.PGPool.Query(context.Background(), query, fmt.Sprintf("%s%%", searchStr), limit, offset)
	if err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	societies, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByPos[societyPresenter])
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	return c.JSON(&fiber.Map{
		"data":    societies,
		"error":   nil,
		"success": true,
	})
}

func BulkInsertBlocks(c fiber.Ctx) error {
	var body models.BlocksCollector

	if err := c.Bind().Body(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid input",
		})
	}

	if body.SocietyID == 0 || len(body.BlockNames) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "all fields are required",
		})
	}

	err := repository.BulkInsertBlocks(body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
	})
}

func BulkInsertResidences(c fiber.Ctx) error {
	var body models.ResidencesCollector

	if err := c.Bind().Body(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid input",
		})
	}

	if body.SocietyID == 0 || body.BlockID == 0 || len(body.ResidenceNumbers) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "all fields are required",
		})
	}

	err := repository.BulkInsertResidences(body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
	})
}

func InsertResident(c fiber.Ctx) error {
	var resident models.ResidentCollector

	if err := c.Bind().Body(&resident); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid input",
		})
	}

	if resident.ResidenceID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "all fields are required",
		})
	}

	id, err := repository.InsertResident(resident)
	if err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"id": id,
		},
	})
}
