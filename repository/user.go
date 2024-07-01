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

func CheckAuthByID(id int, roleLevel int) (bool, error) {
	var exists bool

	if roleLevel == 1 {
		query := `SELECT EXISTS(
		SELECT 1
		    FROM users u
		            INNER JOIN residences rc ON rc.id = u.residence_id
		            INNER JOIN societies sc ON sc.id = rc.society_id
		    WHERE u.role_level = 1
		    AND sc.access_revoked_at IS NULL
		    AND u.access_revoked_at IS NULL
		    AND u.id = $1);`
		err := db.PGPool.QueryRow(context.Background(), query, id).Scan(&exists)
		if err != nil {
			return false, err
		}
		return exists, nil
	} else if roleLevel > 1 && roleLevel < 5 {
		query := `SELECT EXISTS(SELECT 1
		    FROM users u
		            INNER JOIN societies sc ON sc.id = u.society_id
		    WHERE u.role_level > 1
		    AND u.role_level <= 5
		    AND sc.access_revoked_at IS NULL
		    AND u.access_revoked_at IS NULL
		    AND u.id = $1);`
		err := db.PGPool.QueryRow(context.Background(), query, id).Scan(&exists)
		if err != nil {
			return false, err
		}
		return exists, nil
	}
	return false, nil
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
