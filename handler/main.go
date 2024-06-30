package handler

import (
	"context"
	"fmt"
	"hime-backend/db"
	"hime-backend/models"
	"hime-backend/repository"
	"hime-backend/utility"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

func BulkInsertResidentUser(c fiber.Ctx) error {
	var body []models.ResidentUserCollector

	if err := c.Bind().Body(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid input",
		})
	}

	err := repository.BulkInsertResidentUsers(body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
	})
}

func RequestOTP(c fiber.Ctx) error {
	var body models.AuthCollector
	if err := c.Bind().Body(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid input",
		})
	}

	if body.PhoneNumber == nil && body.Email == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "either email or phone number is required",
		})
	}

	otp, secret, err := utility.GenerateVerificationCode()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "unable to generate OTP",
		})
	}

	expiresAt := time.Now().Add(5 * time.Minute)
	if err := repository.StoreAuthSecret(body.PhoneNumber, body.Email, secret, expiresAt); err != nil {
		fmt.Println("err", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "something went wrong, please try again",
		})
	}

	return c.JSON(&fiber.Map{
		"data":    otp,
		"error":   nil,
		"success": true,
	})
}

func VerifyOTP(c fiber.Ctx) error {
	var body struct {
		OTP         string  `json:"otp"`
		PhoneNumber *string `json:"phone_number"`
		Email       *string `json:"email"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid input",
		})
	}
	if body.OTP == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "OTP not provided",
		})
	}

	if body.PhoneNumber == nil && body.Email == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "either email or phone number is required",
		})
	}

	secret, authSecretId, err := repository.GetAuthSecret(body.PhoneNumber, body.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "OTP expired",
		})
	}

	isValid, err := utility.VerifyOTP(body.OTP, secret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "unable to verify OTP",
		})
	}
	if !isValid {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "invalid OTP",
		})
	}

	go repository.MarkAuthSecretAsUsed(authSecretId)

	user, err := repository.GetUserByPhoneNumberOrEmail(body.PhoneNumber, body.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	token, err := utility.CreateJWTToken(utility.JWTAuthClaims{UserID: user.ID, RoleLevel: user.RoleLevel})

	return c.JSON(&fiber.Map{
		"data": &fiber.Map{
			"isValid": isValid,
			"token":   token,
		},
		"error":   nil,
		"success": true,
	})
}

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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	cities, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByPos[cityPresenter])
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
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
	WHERE s.access_revoked_at is null
	  	and lower(s.name) LIKE $1
	LIMIT $2 OFFSET $3;`
	rows, err := db.PGPool.Query(context.Background(), query, fmt.Sprintf("%s%%", searchStr), limit, offset)
	if err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	societies, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByPos[societyPresenter])
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
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

func CreateVisit(c fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "!! failed to parse formdata",
		})
	}

	files := form.File["visitor_photo"]

	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "!! no file uploaded",
		})
	}

	visitorPhoto := files[0]
	file, err := visitorPhoto.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "!! failed to read file",
		})
	}

	utility.UploadFileToS3(file, visitorPhoto, "visitor_photos")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "!! failed to read file",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
	})
}
