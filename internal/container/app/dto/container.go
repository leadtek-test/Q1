package dto

type CreateUserResponse struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}
