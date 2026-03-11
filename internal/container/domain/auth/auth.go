package auth

import "time"

type Claims struct {
	UserID    uint
	Username  string
	ExpiresAt time.Time
}
