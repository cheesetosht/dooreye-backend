package models

type ResidentUserCollector struct {
	ResidenceID int32   `json:"residence_id"`
	Name        string  `json:"name"`
	PhoneNumber *string `json:"phone_number"`
	Email       *string `json:"email"`
}

type AuthCollector struct {
	PhoneNumber *string `json:"phone_number"`
	Email       *string `json:"email"`
}

type BlocksCollector struct {
	BlockNames []string `json:"block_names"`
	SocietyID  int32    `json:"society_id"`
}

type ResidencesCollector struct {
	ResidenceNumbers []int32 `json:"residence_numbers"`
	SocietyID        int32   `json:"society_id"`
	BlockID          int32   `json:"block_id"`
}

type ResidentCollector struct {
	ResidenceID int32 `json:"residence_id"`
	IsPrimary   bool  `json:"is_primary"`
}

type VisitorCollector struct {
	ID            *int32  `json:"id"`
	Name          string  `json:"name"`
	PhoneNumber   string  `json:"phone_number"`
	Photo         *string `json:"photo"`
	Purpose       string  `json:"purpose"`
	IsPreapproved *bool   `json:"is_preapproved"`
}

type ResidenceVisitCollector struct {
	ResidenceID int32                `json:"residence_id"`
	VisitorID   string               `json:"visitor_id"`
	Status      ResidenceVisitStatus `json:"status"`
}
