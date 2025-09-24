package projectsvc

import (
	"context"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/project"
)

type ProjectRepository interface {
	Find(ctx context.Context, id string) (*project.Project, error)
	Store(ctx context.Context, entity *project.Project) error
	Remove(ctx context.Context, ids ...string) error
}

type ProjectService interface {
	List(ctx context.Context, cond project.ListCond, option ListOption) ([]*ViewProject, error)
	Count(ctx context.Context, cond project.CountCond, option CountOption) (*uint64, error)
}

type ListOption struct {
	SortKey ListOptionSortKey
	Size    int64
	Offset  *int64
}

type ListOptionSortKey int

const (
	ListOptionSortKey_CreatedAt_ASC ListOptionSortKey = iota
	ListOptionSortKey_CreatedAt_DESC
)

type CountOption ListOption

type Service interface {
	ShowProject(ctx context.Context, in *ShowProjectIn) (*ShowProjectOut, error)
}

type ShowProjectIn struct{}

type ShowProjectOut struct {
	Item ViewProject
}

type ViewProject struct {
	ProjectID      project.ProjectID
	ProjectUserID  project.UserID
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
