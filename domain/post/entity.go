package post

import "time"

type Post struct {
	PostID        string
	PostProjectID string
	Content       *string
	Visibility    *Visibility
	CreatedTime   time.Time
	UpdatedTime   *time.Time
}

type Visibility string

const (
	Visibility_Public  Visibility = "Public"
	Visibility_Private Visibility = "Private"
)
