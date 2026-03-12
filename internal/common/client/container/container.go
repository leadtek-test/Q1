package container

type CreateContainerRequest struct {
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Command []string          `json:"command"`
	Env     map[string]string `json:"env"`
}

type UpdateContainerStatusRequest struct {
	Action string `json:"action"`
}
