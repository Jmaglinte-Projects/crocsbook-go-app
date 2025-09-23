package poject

import "time"

type Project struct {
	ProjectID      string
	ProjectUserID  string
	Name           string
	Description    *string
	Thumbnail      *string
	Location       *string
	Cost           *int64
	StartDate      *time.Time
	CompletionDate *time.Time
	CreatedTime    time.Time
	UpdatedTime    *time.Time
}
