package dto

import "time"

type ContainerResponse struct {
	ID        uint              `json:"id"`
	UserID    uint              `json:"user_id"`
	Name      string            `json:"name"`
	Image     string            `json:"image"`
	Command   []string          `json:"command"`
	Env       map[string]string `json:"env"`
	RuntimeID string            `json:"runtime_id"`
	Status    string            `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type DeleteContainerResponse struct {
	Deleted bool `json:"deleted"`
}

type CreateContainerJobAcceptedResponse struct {
	JobID string `json:"job_id"`
}

type CreateContainerJobStatusResponse struct {
	JobID        string `json:"job_id"`
	Status       string `json:"status"`
	ContainerID  uint   `json:"container_id,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}
