package domain

type BuildJob struct {
	GitURL    string
	ProjectID int64
	DeployID  int64
	Hostname      string
}
