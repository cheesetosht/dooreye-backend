package model

import (
	"time"

	"github.com/gofrs/uuid"
)

type Visit struct {
	ID           uuid.UUID  `json:"id"`
	ResidenceID  *int64     `json:"residence_id,omitempty"`
	VisitorID    int64      `json:"visitor_id"`
	CheckedInBy  uuid.UUID  `json:"checked_in_by"`
	ApprovedBy   uuid.UUID  `json:"approved_by,omitempty"`
	CheckInTime  time.Time  `json:"check_in_time"`
	CheckOutTime *time.Time `json:"check_out_time,omitempty"`
	Purpose      string     `json:"purpose,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type VisitorType string

const (
	VisitorDelivery    VisitorType = "DELIVERY"
	VisitorMaintenance VisitorType = "MAINTENANCE"
	VisitorGuest       VisitorType = "GUEST"
	VisitorCab         VisitorType = "CAB"
	VisitorStaff       VisitorType = "STAFF"
)

type Visitor struct {
	ID              uuid.UUID   `json:"id"`
	Name            string      `json:"name"`
	Phone           string      `json:"phone"`
	PhotoURL        *string     `json:"photo_url,omitempty"`
	Type            VisitorType `json:"visitor_type"`
	PreApprovedTill *time.Time  `json:"pre_approved_till,omitempty"`
	CreatedBy       uuid.UUID   `json:"created_by"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type VisitWithVisitor struct {
	ID           uuid.UUID  `json:"id"`
	ResidenceID  *int64     `json:"residence_id,omitempty"`
	VisitorID    uuid.UUID  `json:"visitor_id"`
	CheckedInBy  uuid.UUID  `json:"checked_in_by"`
	ApprovedBy   uuid.UUID  `json:"approved_by,omitempty"`
	CheckInTime  time.Time  `json:"check_in_time"`
	CheckOutTime *time.Time `json:"check_out_time,omitempty"`
	Purpose      *string    `json:"purpose,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	Name     string      `json:"name"`
	Phone    string      `json:"phone"`
	PhotoURL *string     `json:"photo_url,omitempty"`
	Type     VisitorType `json:"visitor_type"`
}
