package domain

import (
	"time"
)

type Deployment struct {
	ID          int64
	ProjectID   int64
	HostName    string
	Port        *int
	ContainerID *string
	Status      DeploymentStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type DeployResponse struct {
	DeployID int64
	Url      string
}

// Deployment status - ENUM
type DeploymentStatus string

const (
	StatusPending  DeploymentStatus = "pending"
	StatusBuilding DeploymentStatus = "building"
	StatusRunning  DeploymentStatus = "running"
	StatusFailed   DeploymentStatus = "failed"
	StatusStopped  DeploymentStatus = "stopped"
)
