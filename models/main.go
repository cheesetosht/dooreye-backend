package models

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
)

type AgentRole string

const (
	AgentRoleAdmin    AgentRole = "admin"
	AgentRoleManager  AgentRole = "manager"
	AgentRoleSecurity AgentRole = "security"
)

func (e *AgentRole) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = AgentRole(s)
	case string:
		*e = AgentRole(s)
	default:
		return fmt.Errorf("!! unsupported scan type for AgentRole: %T", src)
	}
	return nil
}

type NullAgentRole struct {
	AgentRole AgentRole
	Valid     bool // Valid is true if AgentRole is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullAgentRole) Scan(value interface{}) error {
	if value == nil {
		ns.AgentRole, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.AgentRole.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullAgentRole) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.AgentRole), nil
}

type VisitStatus string

const (
	VisitStatusAccepted        VisitStatus = "accepted"
	VisitStatusRejected        VisitStatus = "rejected"
	VisitStatusPreApproved     VisitStatus = "pre-approved"
	VisitStatusSecuritycleared VisitStatus = "security cleared"
)

func (e *VisitStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = VisitStatus(s)
	case string:
		*e = VisitStatus(s)
	default:
		return fmt.Errorf("!! unsupported scan type for VisitStatus: %T", src)
	}
	return nil
}

type NullVisitStatus struct {
	VisitStatus VisitStatus
	Valid       bool // Valid is true if VisitStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullVisitStatus) Scan(value interface{}) error {
	if value == nil {
		ns.VisitStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.VisitStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullVisitStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.VisitStatus), nil
}

type Agent struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Mobile    string    `json:"mobile"`
	SocietyID int32     `json:"society_id"`
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

type Resident struct {
	ID          string       `json:"id"`
	ResidenceID int32        `json:"residence_id"`
	IsPrimary   bool         `json:"is_primary"`
	IsValid     bool         `json:"is_valid"`
	CreatedAt   sql.NullTime `json:"created_at"`
	UpdatedAt   sql.NullTime `json:"updated_at"`
}

type Society struct {
	ID            int32        `json:"id"`
	Name          string       `json:"name"`
	Developer     string       `json:"developer"`
	MaxResidences int32        `json:"max_residences"`
	CityID        int32        `json:"city_id"`
	IsValid       bool         `json:"is_valid"`
	CreatedAt     sql.NullTime `json:"created_at"`
	UpdatedAt     sql.NullTime `json:"updated_at"`
}

type Visit struct {
	ID          int32        `json:"id"`
	ResidenceID int32        `json:"residence_id"`
	Status      VisitStatus  `json:"status"`
	ArrivalTime sql.NullTime `json:"arrival_time"`
	ExitTime    sql.NullTime `json:"exit_time"`
}

type Visitor struct {
	ID        uuid.UUID      `json:"id"`
	Name      string         `json:"name"`
	Mobile    string         `json:"mobile"`
	Photo     sql.NullString `json:"photo"`
	CreatedAt sql.NullTime   `json:"created_at"`
	UpdatedAt sql.NullTime   `json:"updated_at"`
}
