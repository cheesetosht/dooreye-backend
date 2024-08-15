package repository

import (
	"context"
	"fmt"
	"hime-backend/db"
	"hime-backend/models"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

func StoreAuthSecret(phoneNumber, email *string, secret string, expiresAt time.Time) error {
	var (
		id         int
		identifier string
		query      string
	)

	if phoneNumber != nil {
		identifier = *phoneNumber
		query = `INSERT INTO auth_secrets (phone_number, secret, expires_at) VALUES ($1, $2, $3) returning id;`
	} else if email != nil {
		identifier = *email
		query = `INSERT INTO auth_secrets (email, secret, expires_at) VALUES ($1, $2, $3) returning id;`
	}
	err := db.PGPool.QueryRow(context.Background(), query, identifier, secret, expiresAt).Scan(&id)
	if err != nil {
		return fmt.Errorf("!! failed to store auth secret: %s", err)
	}
	return nil
}

func GetAuthSecret(phoneNumber, email *string) (string, int, error) {
	var (
		id            int
		secret        string
		identifierKey string
		identifier    string
	)
	if phoneNumber != nil {
		identifierKey = "phone_number"
		identifier = *phoneNumber
	} else if email != nil {
		identifierKey = "email"
		identifier = *email
	}
	query := "SELECT secret, id FROM auth_secrets WHERE " + identifierKey + " = $1 AND is_used = FALSE AND expires_at > NOW() ORDER BY id DESC LIMIT 1;"
	err := db.PGPool.QueryRow(context.Background(), query, identifier).Scan(&secret, &id)
	if err != nil {
		fmt.Println("OTP", err)
		return "", 0, fmt.Errorf("!! failed to fetch auth secret: %s", err)
	}
	return secret, id, nil
}

func MarkAuthSecretAsUsed(id int) error {
	query := `UPDATE auth_secrets SET is_used = TRUE WHERE id = $1 AND expires_at > NOW() AND is_used = FALSE RETURNING id;`
	db.PGPool.QueryRow(context.Background(), query, id)
	log.Println("> delete from auth_secrets, id: ", id)
	return nil
}

func GetUserByID(id int32) (models.User, error) {
	var (
		user models.User
	)

	query := "SELECT * FROM users WHERE id = $1 AND access_revoked_at IS NULL;"
	err := db.PGPool.QueryRow(context.Background(), query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PhoneNumber,
		&user.ResidenceID,
		&user.SocietyID,
		&user.RoleLevel,
		&user.AccessRevokedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	return user, err
}

func GetUserByPhoneNumberOrEmail(phoneNumber, email *string) (models.User, error) {
	var (
		identifierKey, identifier string
		user                      models.User
	)
	if phoneNumber != nil {
		identifierKey = "phone_number"
		identifier = *phoneNumber
	} else if email != nil {
		identifierKey = "email"
		identifier = *email
	}
	query := "SELECT * FROM users WHERE " + identifierKey + " = $1 AND access_revoked_at IS NULL;"
	err := db.PGPool.QueryRow(context.Background(), query, identifier).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PhoneNumber,
		&user.ResidenceID,
		&user.SocietyID,
		&user.RoleLevel,
		&user.AccessRevokedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	return user, err
}

var getUserInfoQuery = `SELECT u.id,
       u.name,
       u.email,
       u.phone_number,
       u.residence_id,
       u.society_id,
       u.created_at,
       rc.number residence_number,
       bl.name block,
       sc.name society_name,
       ct.name city_name,
       u.role_level
FROM users u
         LEFT JOIN residences rc ON rc.id = u.residence_id
         LEFT JOIN societies sc ON sc.id = rc.society_id
         LEFT JOIN blocks bl ON bl.id = rc.block_id
         LEFT JOIN cities ct ON ct.id = sc.city_id
WHERE sc.access_revoked_at IS NULL
  AND u.access_revoked_at IS NULL
  AND u.id = $1 AND u.role_level >= $2
LIMIT 1`

func GetUserInfoByIDAndRoleLevel(id int, roleLevel int) (*models.UserInfo, error) {
	var i models.UserInfo
	err := db.PGPool.QueryRow(context.Background(), getUserInfoQuery, id, roleLevel).Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.PhoneNumber,
		&i.ResidenceID,
		&i.SocietyID,
		&i.CreatedAt,
		&i.ResidenceNumber,
		&i.Block,
		&i.SocietyName,
		&i.CityName,
		&i.RoleLevel,
	)

	if err != nil {
		return nil, fmt.Errorf("!! failed to collecting user info: %w", err)
	}
	return &i, err
}

func BulkInsertResidentUsers(data []models.ResidentUserCollector) error {
	context := context.Background()
	conn, err := db.PGPool.Acquire(context)
	if err != nil {
		return fmt.Errorf("!! failed to acquire db conn.: %w", err)
	}

	defer conn.Release()

	tx, err := conn.Begin(context)
	if err != nil {
		return fmt.Errorf("!! failed to start txn: %w", err)
	}

	batch := &pgx.Batch{}

	for _, user := range data {
		var identifierKey, identifier string
		if user.PhoneNumber != nil {
			identifierKey = "phone_number"
			identifier = *user.PhoneNumber
		} else if user.Email != nil {
			identifierKey = "email"
			identifier = *user.Email
		}
		query := "INSERT INTO users (" + identifierKey + ", name, residence_id, role_level) VALUES ($1, $2, $3, 1)"
		batch.Queue(query, identifier, user.Name, user.ResidenceID)
	}

	br := tx.SendBatch(context, batch)
	err = br.Close()
	if err != nil {
		tx.Rollback(context)
		return fmt.Errorf("!! failed to execute batch insert: %w", err)
	}

	err = tx.Commit(context)
	if err != nil {
		return fmt.Errorf("!! failed to commit transaction: %w", err)
	}

	return nil
}
