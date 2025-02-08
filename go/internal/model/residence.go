package model

import "time"

type City struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Society struct {
	ID        int64     `json:"id"`
	CityID    int64     `json:"city_id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
}

type Block struct {
	ID        int64  `json:"id"`
	SocietyID int64  `json:"society_id"`
	Name      string `json:"name"`
}

type Residence struct {
	ID      int64  `json:"id"`
	BlockID int64  `json:"block_id"`
	Number  string `json:"number"`
	Floor   int    `json:"floor"`
}
