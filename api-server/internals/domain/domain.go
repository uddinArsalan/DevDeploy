package domain

type DeployResponse struct {
	Url string
}

type BuildJob struct{
	GitURL string
	ProjectID string 
	Slug string
}