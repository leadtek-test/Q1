package container

type RegisterRequest struct {
	Username string `json:"username" form:"required"`
	Password string `json:"password" form:"required"`
}
