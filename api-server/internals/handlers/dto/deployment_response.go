package dto

import (
	"time"

	"github.com/uddinArsalan/devdeploy/internals/domain"
)

type DeployResponse struct {
	DeployID int64  `json:"deploy_id"`
	URL      string `json:"url"`
}

type DeploymentResponse struct {
	ID        int64                   `json:"id"`
	HostName  string                  `json:"hostname"`
	Status    domain.DeploymentStatus `json:"status"`
	CreatedAt time.Time               `json:"created_at"`
	UpdatedAt time.Time               `json:"updated_at"`
}
