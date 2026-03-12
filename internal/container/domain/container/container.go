package container

import "time"

type Status string

const (
	StatusCreated Status = "created"
	StatusRunning Status = "running"
	StatusStopped Status = "stopped"
)

type Container struct {
	ID        uint
	UserID    uint
	Name      string
	Image     string
	Command   []string
	Env       map[string]string
	RuntimeID string
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateSpec struct {
	Name    string
	Image   string
	Command []string
	Env     map[string]string
}
