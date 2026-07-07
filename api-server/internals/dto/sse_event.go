package dto

import "github.com/uddinArsalan/devdeploy/internals/domain"

type DeploymentEvent struct {
	Type   string                  `json:"type"`
	Status domain.DeploymentStatus `json:"status,omitempty"`
	Log    string                  `json:"log,omitempty"`
}
