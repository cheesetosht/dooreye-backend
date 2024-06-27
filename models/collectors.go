package models

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
	Name   string `json:"name"`
	Mobile string `json:"mobile"`
	Photo  string `json:"photo"`
}

type ResidenceVisitCollector struct {
	ResidenceID int32                `json:"residence_id"`
	VisitorID   string               `json:"visitor_id"`
	Status      ResidenceVisitStatus `json:"status"`
}
