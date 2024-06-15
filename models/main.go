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
	ID        uuid.UUID
	Name      string
	Mobile    string
	SocietyID int32
}

type Block struct {
	ID        int32
	Name      string
	SocietyID int32
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
	ID        int32
	Number    int32
	SocietyID int32
	BlockID   sql.NullInt32
	CreatedAt sql.NullTime
	UpdatedAt sql.NullTime
}

type Resident struct {
	ID          string
	ResidenceID int32
	IsPrimary   bool
	IsActive    bool
	CreatedAt   sql.NullTime
	UpdatedAt   sql.NullTime
}

type Society struct {
	ID            int32
	Name          string
	Developer     string
	MaxResidences int32
	CityID        int32
	IsActive      bool
	CreatedAt     sql.NullTime
	UpdatedAt     sql.NullTime
}

type Visit struct {
	ID          int32
	ResidenceID int32
	Status      VisitStatus
	ArrivalTime sql.NullTime
	ExitTime    sql.NullTime
}

type Visitor struct {
	ID        uuid.UUID
	Name      string
	Mobile    string
	Photo     sql.NullString
	CreatedAt sql.NullTime
	UpdatedAt sql.NullTime
}
