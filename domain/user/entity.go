package user

import "time"

type User struct {
	UserID     string
	Email      string
	Gender     string
	ProfileURL *string
	Nickname   *string
	Username   *string
	Password   *string
	CreatedAt  time.Time
	UpdatedAt  *time.Time
}
