package dto

import "time"

type CreateUserResponse struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}

type LoginUserResponse struct {
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
