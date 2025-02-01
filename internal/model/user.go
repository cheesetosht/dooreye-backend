package model

import (
	"time"

	"github.com/gofrs/uuid"
)

type UserRole string

const (
	RoleAdmin          UserRole = "ADMIN"
	RoleSocietyManager UserRole = "SOCIETY_MANAGER"
	RoleSecurity       UserRole = "SECURITY"
	RoleOwner          UserRole = "OWNER"
	RoleResident       UserRole = "RESIDENT"
)

type User struct {
	ID          uuid.UUID  `json:"id"`
	AccessCode  string     `json:"access_code,omitempty"`
	Role        UserRole   `json:"role"`
	Name        string     `json:"name"`
	ResidenceID int64      `json:"residence_id,omitempty"`
	SocietyID   int64      `json:"society_id,omitempty"`
	IsActive    bool       `json:"is_active"`
	DeviceID    string     `json:"device_id"`
	ActivatedBy uuid.UUID  `json:"activated_by,omitempty"`
	ActivatedAt *time.Time `json:"activated_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
