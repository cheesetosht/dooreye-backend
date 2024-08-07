package handler

import (
	"context"
	"fmt"
	"hime-backend/db"
	"hime-backend/models"
	"hime-backend/repository"
	"hime-backend/utility"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"firebase.google.com/go/messaging"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

func BulkCreateResidentUser(c fiber.Ctx) error {
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "something went wrong, please try again",
		})
	}

	if err := utility.SendSMS(*body.PhoneNumber, "Your secure code access your HIME account is "+otp); err != nil {
		fmt.Println(err)
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

	token, err := utility.CreateJWTToken(utility.JWTAuthClaims{UserID: user.ID, RoleLevel: user.RoleLevel, ResidenceID: *user.ResidenceID})

	return c.JSON(&fiber.Map{
		"data": &fiber.Map{
			"isValid": isValid,
			"token":   token,
		},
		"error":   nil,
		"success": true,
	})
}

func VerifyToken(c fiber.Ctx) error {
	localsUserInfo := c.Locals("user_info")

	userInfo, ok := localsUserInfo.(*models.UserInfo)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "invalid user ID",
		})
	}

	return c.JSON(&fiber.Map{
		"data": &fiber.Map{
			"user_info": userInfo,
		},
	})
}

func CreateCity(c fiber.Ctx) error {
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

func FetchCities(c fiber.Ctx) error {
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

func CreateSociety(c fiber.Ctx) error {
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

func FetchSocieties(c fiber.Ctx) error {
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

func BulkCreateBlocks(c fiber.Ctx) error {
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

func BulkCreateResidences(c fiber.Ctx) error {
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

func CreateVisitByVisitorID(c fiber.Ctx) error {
	var body struct {
		VisitorID   int `json:"visitor_id"`
		ResidenceID int `json:"residence_id"`
	}

	if err := c.Bind().Body(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid input",
		})
	}

	visitor, err := repository.GetVisitorByID(body.VisitorID)
	fmt.Println(body.VisitorID)
	if err != nil {
		return c.Status(fiber.StatusNotAcceptable).JSON(fiber.Map{
			"error": "visitor does not exist",
		})
	}

	var status *models.ResidenceVisitStatus
	if visitor.IsPreapproved {
		preapprovedStatus := models.ResidenceVisitStatusPreApproved
		status = &preapprovedStatus
	}

	visit, err := repository.InsertResidenceVisit(body.ResidenceID, body.VisitorID, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("something went wrong: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"data": &fiber.Map{
			"visit":   visit,
			"visitor": visitor,
		},
		"success": true,
	})

}

func CreateVisitForNewVisitor(c fiber.Ctx) error {
	residenceIDStr := c.FormValue("residence_id")
	name := c.FormValue("name")
	phoneNumber := c.FormValue("phone_number")
	purpose := c.FormValue("purpose")
	residenceID, err := strconv.Atoi(residenceIDStr)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid residence_id",
		})
	}

	var photoURL *string

	file, err := c.FormFile("visitor_photo")
	if err == nil && file != nil {
		var filename string

		if name == "" {
			filename = phoneNumber
		} else {
			filename = name
		}
		filename += filepath.Ext(file.Filename)
		filename = strings.ReplaceAll(filename, " ", "_")

		s3Path := fmt.Sprintf("visitors/%s", filename)

		s3URL, err := utility.UploadFileToS3(file, s3Path, os.Getenv("AWS_S3_BUCKET_NAME"))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("unable to upload image: %v", err),
			})
		}
		photoURL = &s3URL
	} else if err != nil && err != http.ErrMissingFile {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "error processing file upload",
		})
	}

	visitor, err := repository.InsertVisitor(models.VisitorCollector{
		Name:        name,
		PhoneNumber: phoneNumber,
		Photo:       photoURL,
		Purpose:     purpose,
	}, false)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("something went wrong: %v", err),
		})
	}

	var status *models.ResidenceVisitStatus
	if *visitor.IsPreapproved {
		preapprovedStatus := models.ResidenceVisitStatusPreApproved
		status = &preapprovedStatus
	}

	visit, err := repository.InsertResidenceVisit(residenceID, int(*visitor.ID), status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("something went wrong: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"data": &fiber.Map{
			"visit":   visit,
			"visitor": visitor,
		},
		"success": true,
	})

}

func CreateVisitorFromResidence(c fiber.Ctx) error {
	name := c.FormValue("name")
	phoneNumber := c.FormValue("phone_number")
	purpose := c.FormValue("purpose")

	visitor, err := repository.InsertVisitor(models.VisitorCollector{
		Name:        name,
		PhoneNumber: phoneNumber,
		Purpose:     purpose,
	}, true)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("something went wrong: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"data": &fiber.Map{
			"visitor": visitor,
		},
		"success": true,
	})

}

func GetVisitor(c fiber.Ctx) error {
	phoneNumber := c.Query("phone_number")

	if phoneNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "atleast 4 digits of phone number is required",
		})
	}

	visitor, err := repository.GetVisitorByMobile(phoneNumber)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "no visitor found with this phone number",
		})
	}

	return c.JSON(&fiber.Map{
		"data":    visitor,
		"error":   nil,
		"success": true,
	})
}

func ChangeVisitStatus(c fiber.Ctx) error {
	var body struct {
		Status models.ResidenceVisitStatus `json:"status"`
	}

	if err := c.Bind().Body(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid input",
		})
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id parameter",
		})
	}

	repository.UpdateResidenceVisitStatus(id, &body.Status)

	return c.JSON(fiber.Map{
		"success": true,
	})
}

func SendPushNotification(c fiber.Ctx) error {
	var data struct {
		Token string `json:"token"`
		Title string `json:"title"`
		Body  string `json:"body"`
	}

	if err := c.Bind().Body(&data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse JSON",
		})
	}

	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: data.Title,
			Body:  data.Body,
		},
		Token: data.Token,
	}

	client := utility.GetFirebaseMessagingClient()

	response, err := client.Send(c.Context(), message)
	if err != nil {
		log.Printf("!! error sending push notification: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to send notification",
		})
	}

	return c.JSON(fiber.Map{
		"message":  "push notification sent successfully",
		"response": response,
	})
}
