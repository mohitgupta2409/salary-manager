package model

import "time"

type Country struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Currency  string    `json:"currency"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
