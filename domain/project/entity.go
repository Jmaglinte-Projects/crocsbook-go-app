package project

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ProjectID      ProjectID
	ProjectUserID  UserID
	Name           string
	Description    *string
	ThumbnailKey   *string
	Location       *string
	Cost           *int64
	StartDate      *time.Time
	CompletionDate *time.Time
	CreatedTime    time.Time
	UpdatedTime    *time.Time

	ThumbnailSet ThumbnailSet
}

type ProjectID string

func NewProjectID() (ProjectID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return ProjectID(id.String()), nil
}

type UserID string

type ThumbnailSet struct {
	ContentType string
	Content     []byte
}

type ProjectLike struct {
	ProjectID   ProjectID
	UserID      UserID
	CreatedTime time.Time
}

type ListCond struct {
	ProjectID  *ProjectID
	ProjectIDs []ProjectID
}

type CountCond ListCond

type ProjectLikeListCond struct {
	ProjectID *ProjectID
	UserID    *UserID
}

type ProjectLikeCountCond ProjectLikeListCond
