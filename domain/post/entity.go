package post

import "time"

type Post struct {
	PostID        PostID
	PostProjectID ProjectID
	Content       *string
	Visibility    *Visibility
	CreatedTime   time.Time
	UpdatedTime   *time.Time
}

type PostID string

type ProjectID string

type Visibility string

const (
	Visibility_Public  Visibility = "Public"
	Visibility_Private Visibility = "Private"
)
