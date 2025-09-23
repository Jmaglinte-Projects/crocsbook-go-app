package project

import "time"

type Project struct {
	ProjectID      ProjectID
	ProjectUserID  UserID
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

type ProjectID string

type UserID string

type ListCond struct {
	ProjectID  *ProjectID
	ProjectIDs []ProjectID
}

type CountCond ListCond
