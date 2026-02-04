package entity

import "time"

type User struct {
	ID        int64     `json:"user_id"`
	Name      string    `json:"name"`
	UserName  string	`json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // "-" agar password tidak ikut terkirim saat return JSON ke frontend
	Role 	  string 	`json:"role"`
	CreatedAt time.Time `json:"created_at"`
}