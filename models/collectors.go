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

type VisitorCollector struct {
	Name   string
	Mobile string
	Photo  string
}

type VisitCollector struct {
	ResidenceID string      `json:"residence_id"`
	VisitorID   string      `json:"visitor_id"`
	Status      VisitStatus `json:"status"`
}
