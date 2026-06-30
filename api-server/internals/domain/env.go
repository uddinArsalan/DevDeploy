package domain

import "time"

type Env struct {
	ID             int64
	Key            string
	EncryptedValue []byte
	ProjectID      int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
