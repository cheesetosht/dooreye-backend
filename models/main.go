package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type ResidenceVisitStatus string

const (
	ResidenceVisitStatusAccepted        ResidenceVisitStatus = "accepted"
	ResidenceVisitStatusRejected        ResidenceVisitStatus = "rejected"
	ResidenceVisitStatusPreApproved     ResidenceVisitStatus = "pre-approved"
	ResidenceVisitStatusSecuritycleared ResidenceVisitStatus = "security cleared"
)

func (e *ResidenceVisitStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ResidenceVisitStatus(s)
	case string:
		*e = ResidenceVisitStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for ResidenceVisitStatus: %T", src)
	}
	return nil
}

type NullResidenceVisitStatus struct {
	ResidenceVisitStatus ResidenceVisitStatus
	Valid                bool // Valid is true if ResidenceVisitStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullResidenceVisitStatus) Scan(value interface{}) error {
	if value == nil {
		ns.ResidenceVisitStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ResidenceVisitStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullResidenceVisitStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ResidenceVisitStatus), nil
}

type Block struct {
	ID        int32  `json:"id"`
	Name      string `json:"name"`
	SocietyID int32  `json:"society_id"`
}

type City struct {
	ID      int32   `json:"id"`
	Name    string  `json:"name"`
	State   string  `json:"state"`
	Country string  `json:"country"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}

type Residence struct {
	ID        int32      `json:"id"`
	Number    int32      `json:"number"`
	SocietyID int32      `json:"society_id"`
	BlockID   *int       `json:"block_id"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type ResidenceVisit struct {
	ID          int32                `json:"id"`
	ResidenceID int32                `json:"residence_id"`
	Status      ResidenceVisitStatus `json:"status"`
	ArrivalTime *time.Time           `json:"arrival_time"`
	ExitTime    *time.Time           `json:"exit_time"`
}

type Society struct {
	ID              int32      `json:"id"`
	Name            string     `json:"name"`
	Developer       string     `json:"developer"`
	MaxResidences   int32      `json:"max_residences"`
	CityID          int32      `json:"city_id"`
	AccessRevokedAt *time.Time `json:"access_revoked_at"`
	CreatedAt       *time.Time `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
}

type User struct {
	ID              int32      `json:"id"`
	Name            *string    `json:"name"`
	Email           *string    `json:"email"`
	PhoneNumber     *string    `json:"phone_number"`
	ResidenceID     *int       `json:"residence_id"`
	SocietyID       *int       `json:"society_id"`
	RoleLevel       int32      `json:"role_level"`
	AccessRevokedAt *time.Time `json:"access_revoked_at"`
	CreatedAt       *time.Time `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
}

type UserRole struct {
	Level int32
	Role  *string
}

type Visitor struct {
	ID            int32      `json:"id"`
	Name          string     `json:"name"`
	PhoneNumber   string     `json:"phone_number"`
	Photo         *string    `json:"photo"`
	Purpose       string     `json:"purpose"`
	IsPreapproved bool       `json:"is_preapproved"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
}
