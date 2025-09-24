package user

import "time"

type User struct {
	UserID      UserID
	Email       string
	Gender      Gender
	ProfileURL  *string
	Nickname    *string
	Username    *string
	Password    *string
	CreatedTime time.Time
	UpdatedTime *time.Time
}

type UserID string

type Gender string

const (
	Gender_Male   Gender = "Male"
	Gender_Female Gender = "Female"
)

type ListCond struct {
	UserID  *UserID
	UserIDs []UserID
}

type CountCond ListCond
