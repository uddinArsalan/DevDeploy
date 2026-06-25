package dto

type DeployResponse struct {
	DeployID int64  `json:"deploy_id"`
	URL      string `json:"url"`
}
