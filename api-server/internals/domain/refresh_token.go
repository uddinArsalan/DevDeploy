package domain

import "time"

type RefreshToken struct {
	ID                int64
	UserID            int64
	TokenHash         string
	ReplacedByTokenID *int64
	RevokedAt         *time.Time
	ExpiresAt         time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
