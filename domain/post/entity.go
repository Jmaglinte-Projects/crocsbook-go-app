package post

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	PostID        PostID
	PostProjectID ProjectID
	Content       *string
	Visibility    *Visibility
	CreatedTime   time.Time
	UpdatedTime   *time.Time

	MediaSets []MediaSet
}

type PostID string

func NewPostID() (PostID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return PostID(id.String()), nil
}

type ProjectID string

type Visibility string

const (
	Visibility_Public  Visibility = "Public"
	Visibility_Private Visibility = "Private"
)

type MediaSet struct {
	ContentType string
	Content     []byte
}

type ListCond struct {
	PostID         *PostID
	PostIDs        []PostID
	PostProjectID  *ProjectID
	PostProjectIDs []ProjectID
}

type CountCond ListCond
