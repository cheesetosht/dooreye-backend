package models

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
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
	ID        int32         `json:"id"`
	Number    int32         `json:"number"`
	SocietyID int32         `json:"society_id"`
	BlockID   sql.NullInt32 `json:"block_id"`
	CreatedAt sql.NullTime  `json:"created_at"`
	UpdatedAt sql.NullTime  `json:"updated_at"`
}

type ResidenceVisit struct {
	ID          int32                `json:"id"`
	ResidenceID int32                `json:"residence_id"`
	Status      ResidenceVisitStatus `json:"status"`
	ArrivalTime sql.NullTime         `json:"arrival_time"`
	ExitTime    sql.NullTime         `json:"exit_time"`
}

type Society struct {
	ID              int32        `json:"id"`
	Name            string       `json:"name"`
	Developer       string       `json:"developer"`
	MaxResidences   int32        `json:"max_residences"`
	CityID          int32        `json:"city_id"`
	AccessRevokedAt sql.NullTime `json:"access_revoked_at"`
	CreatedAt       sql.NullTime `json:"created_at"`
	UpdatedAt       sql.NullTime `json:"updated_at"`
}

type User struct {
	ID              int32          `json:"id"`
	Name            sql.NullString `json:"name"`
	Email           sql.NullString `json:"email"`
	Mobile          sql.NullString `json:"mobile"`
	ResidenceID     sql.NullInt32  `json:"residence_id"`
	SocietyID       sql.NullInt32  `json:"society_id"`
	RoleLevel       int32          `json:"role_level"`
	AccessRevokedAt sql.NullTime   `json:"access_revoked_at"`
	CreatedAt       sql.NullTime   `json:"created_at"`
	UpdatedAt       sql.NullTime   `json:"updated_at"`
}

type UserRole struct {
	Level int32
	Role  sql.NullString
}

type Visitor struct {
	ID        int32          `json:"id"`
	Name      string         `json:"name"`
	Mobile    string         `json:"mobile"`
	Photo     sql.NullString `json:"photo"`
	CreatedAt sql.NullTime   `json:"created_at"`
	UpdatedAt sql.NullTime   `json:"updated_at"`
}
