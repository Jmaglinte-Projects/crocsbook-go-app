package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserID      UserID
	Email       string
	Gender      Gender
	ProfileKey  *string
	Nickname    *string
	Username    *string
	Password    *string
	CreatedTime time.Time
	UpdatedTime *time.Time

	ProfileURL *string
	ImageSet   *ImageSet
}

type UserID string

func NewUserID() (UserID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return UserID(id.String()), nil
}

type Gender string

const (
	Gender_Male   Gender = "Male"
	Gender_Female Gender = "Female"
)

type ImageSet struct {
	ContentType string
	Content     []byte
}

type ListCond struct {
	Email   *string
	UserID  *UserID
	UserIDs []UserID
}

type CountCond ListCond
