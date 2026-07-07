package dto

type DeployReqDTO struct {
	ProjectID int64 `json:"project_id"`
}

type StopDeployReqDTO struct {
	DeployID int64 `json:"deploy_id"`
}