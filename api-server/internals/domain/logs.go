package domain

type LogType string

const (
	Status LogType = "status"
	Log    LogType = "log"
)

type LogEvent struct {
	ID   string
	Type LogType
	Data interface{}
}
