package domain

import "time"

type Project struct {
	ID        int64
	Name      string
	GitUrl    string
	CreatedAt time.Time
}

type CreateProjectResult struct {
	ProjectID int64
}
