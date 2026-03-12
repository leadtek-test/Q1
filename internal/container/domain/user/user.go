package user

import "time"

type User struct {
	ID             uint
	Username       string
	PasswordHashed string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
