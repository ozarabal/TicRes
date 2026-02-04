package entity

import "time"

type Event struct{
	ID		int64	`json:"event_id"`
	Name	string 	`json:"name"`
	Location	string	`json:"location"`
	Date      time.Time `json:"date"`
	Capacity  int       `json:"capacity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

