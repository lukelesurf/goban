package model

import "time"

type Column struct {
	ID        string    `json:"id"`
	BoardID   string    `json:"board_id"`
	Name      string    `json:"name"`
	Position  int       `json:"position"`
	WIPLimit  int       `json:"wip_limit"`
	CreatedAt time.Time `json:"created_at"`
}
