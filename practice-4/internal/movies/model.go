package movies

import "time"

type Movie struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Genre     string    `json:"genre"`
	Budget    int64     `json:"budget"`
	CreatedAt time.Time `json:"created_at"`
}