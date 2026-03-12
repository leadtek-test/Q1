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
