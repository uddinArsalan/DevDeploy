package dto

type UserURLReqDTO struct{
	GitURL string `json:"git_url"`
}

type DeployReqDTO struct{
	DeployID string `json:"deploy_id"`
}

type DeployResponse struct {
	Url string `json:"url"`
	DeployID string `json:"deploy_id"`
}