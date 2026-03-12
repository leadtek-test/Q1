package job

import "time"

type CreateContainerJobStatus string

const (
	CreateContainerJobStatusAccepted  CreateContainerJobStatus = "accepted"
	CreateContainerJobStatusCreating  CreateContainerJobStatus = "creating"
	CreateContainerJobStatusSucceeded CreateContainerJobStatus = "succeeded"
	CreateContainerJobStatusFailed    CreateContainerJobStatus = "failed"
)

type CreateContainerJob struct {
	JobID        string
	UserID       uint
	Name         string
	Image        string
	Command      []string
	Env          map[string]string
	Status       CreateContainerJobStatus
	ErrorMessage string
	ContainerID  uint
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateContainerTask struct {
	UserID  uint
	Name    string
	Image   string
	Command []string
	Env     map[string]string
}
