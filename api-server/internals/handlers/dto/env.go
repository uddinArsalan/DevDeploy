package dto

import "time"

type Env struct {
	ID        int64     `json:"id,omitempty"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
