package store

import (
	"context"
	"dooreye-backend/internal/model"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx"
)

var (
	ErrInvalidUserType     = errors.New("either residence_id or society_id is required")
	ErrDuplicateAccessCode = errors.New("access code already exists")
)

type AuthUser struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	SocietyID *int64 `json:"society_id,omitempty"`
	IsActive  bool   `json:"is_active"`
}

type User struct {
	ID          string     `json:"id"`
	AccessCode  string     `json:"access_code"`
	DeviceID    string     `json:"device_id"`
	Name        string     `json:"name"`
	ResidenceID *int64     `json:"residence_id,omitempty"`
	SocietyID   *int64     `json:"society_id,omitempty"`
	Role        string     `json:"role"`
	IsActive    bool       `json:"is_active"`
	ActivatedBy string     `json:"activated_by"`
	ActivatedAt *time.Time `json:"activated_at"`
}

type CreateUserParams struct {
	AccessCode  string
	DeviceID    string
	Name        string
	ResidenceID *int64
	SocietyID   *int64
	Role        model.UserRole
	ActivatedBy string
}

func (db *DB) CreateUser(ctx context.Context, params CreateUserParams) (*User, error) {
	// Validate user type
	if params.ResidenceID == nil && params.SocietyID == nil {
		return nil, ErrInvalidUserType
	}

	// First check if access code exists
	var exists bool
	err := db.pool.QueryRow(ctx, `
        SELECT EXISTS(
            SELECT 1 FROM users WHERE access_code = $1
        )
    `, params.AccessCode).Scan(&exists)

	if err != nil {
		return nil, fmt.Errorf("checking access code: %w", err)
	}

	if exists {
		return nil, ErrDuplicateAccessCode
	}

	tx, err := db.BeginTx(ctx)
	var user User
	tx.QueryRow(ctx, `
            INSERT INTO users (
                access_code, device_id, name, residence_id, society_id,
                role, is_active, activated_by, activated_at
            )
            VALUES ($1, $2, $3, $4, $5, $6, true, $7, NOW())
            RETURNING id, access_code, device_id, name, residence_id,
                      society_id, role, is_active, activated_by, activated_at
        `,
		params.AccessCode,
		params.DeviceID,
		params.Name,
		params.ResidenceID,
		params.SocietyID,
		params.Role,
		params.ActivatedBy,
	).Scan(
		&user.ID,
		&user.AccessCode,
		&user.DeviceID,
		&user.Name,
		&user.ResidenceID,
		&user.SocietyID,
		&user.Role,
		&user.IsActive,
		&user.ActivatedBy,
		&user.ActivatedAt,
	)

	if err != nil {
		if isPgError(err, "unique_violation") {
			return nil, ErrDuplicateAccessCode
		}
		return nil, fmt.Errorf("creating user: %w", err)
	}

	return &user, nil
}

func (db *DB) GetUserByDeviceID(ctx context.Context, deviceID string) (*AuthUser, error) {
	query := `
        SELECT id, role, society_id, is_active
        FROM users
        WHERE device_id = $1
    `

	var user AuthUser
	err := db.pool.QueryRow(ctx, query, deviceID).Scan(
		&user.ID,
		&user.Role,
		&user.SocietyID,
		&user.IsActive,
	)

	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by device ID: %w", err)
	}

	return &user, nil
}
