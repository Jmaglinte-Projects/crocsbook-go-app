package projectsvc

import (
	"context"
	"errors"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/project"
)

type ProjectRepository interface {
	Find(ctx context.Context, id project.ProjectID) (*ViewProject, error)
	Store(ctx context.Context, entity *project.Project) error
	Remove(ctx context.Context, ids ...project.ProjectID) error
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
	ShowProjects(ctx context.Context, in *ShowProjectsIn) (*ShowProjectsOut, error)
	ShowProject(ctx context.Context, in *ShowProjectIn) (*ShowProjectOut, error)
	CreateProject(ctx context.Context, in *CreateProjectIn) (*CreateProjectOut, error)
	UpdateProject(ctx context.Context, in *UpdateProjectIn) (*UpdateProjectOut, error)
	RemoveProject(ctx context.Context, in *RemoveProjectIn) (*RemoveProjectOut, error)
}

type ShowProjectsIn struct{}
type ShowProjectsOut struct {
	Items []*ViewProject
	Total uint64
}

type ShowProjectIn struct {
	ProjectID project.ProjectID
}
type ShowProjectOut struct {
	Item *ViewProject
}

type CreateProjectIn struct {
	ProjectUserID  project.UserID
	Name           string
	Description    *string
	Thumbnail      *string
	Location       *string
	Cost           *int64
	StartDate      *time.Time
	CompletionDate *time.Time
}
type CreateProjectOut struct{}

type UpdateProjectIn struct {
	ProjectID project.ProjectID

	Name           string
	Description    *string
	Thumbnail      *string
	Location       *string
	Cost           *int64
	StartDate      *time.Time
	CompletionDate *time.Time
}
type UpdateProjectOut struct{}

type RemoveProjectIn struct {
	ProjectID project.ProjectID
}
type RemoveProjectOut struct{}

type service struct {
	projectRepo ProjectRepository
	projectSvc  ProjectService
}

func NewService(projectRepo ProjectRepository, projectSvc ProjectService) Service {
	return &service{
		projectRepo: projectRepo,
		projectSvc:  projectSvc,
	}
}

func (s *service) ShowProjects(ctx context.Context, in *ShowProjectsIn) (*ShowProjectsOut, error) {
	cond := &project.ListCond{}
	opt := &ListOption{}

	entities, err := s.projectSvc.List(ctx, *cond, *opt)
	if err != nil {
		return nil, err
	}

	countCond := &project.CountCond{}
	countOpt := &CountOption{}
	count, err := s.projectSvc.Count(ctx, *countCond, *countOpt)
	if err != nil {
		return nil, err
	}

	return &ShowProjectsOut{
		Items: entities,
		Total: *count,
	}, nil
}

func (s *service) ShowProject(ctx context.Context, in *ShowProjectIn) (*ShowProjectOut, error) {
	entity, err := s.projectRepo.Find(ctx, in.ProjectID)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, errors.New("Entity not found")
	}

	return &ShowProjectOut{
		Item: entity,
	}, nil
}

func (s *service) CreateProject(ctx context.Context, in *CreateProjectIn) (*CreateProjectOut, error) {
	now := time.Now()

	id, err := project.NewProjectID()
	if err != nil {
		return nil, err
	}

	entity := &project.Project{}
	entity.ProjectID = id
	entity.ProjectUserID = in.ProjectUserID
	entity.Name = in.Name
	entity.Description = in.Description
	entity.Thumbnail = in.Thumbnail
	entity.Location = in.Location
	entity.Cost = in.Cost
	entity.StartDate = in.StartDate
	entity.CompletionDate = in.CompletionDate
	entity.CreatedTime = now

	if err = s.projectRepo.Store(ctx, entity); err != nil {
		return nil, err
	}

	return &CreateProjectOut{}, nil
}

func (s *service) UpdateProject(ctx context.Context, in *UpdateProjectIn) (*UpdateProjectOut, error) {
	now := time.Now()

	entity, err := s.projectRepo.Find(ctx, in.ProjectID)
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, errors.New("Entity not found")
	}

	entity.Name = in.Name
	entity.Description = in.Description

	// TODO: thumbnail should not be string type here.. refactor later
	entity.Thumbnail = in.Thumbnail

	entity.Location = in.Location
	entity.Cost = in.Cost
	entity.StartDate = in.StartDate
	entity.CompletionDate = in.CompletionDate
	entity.UpdatedTime = &now

	if err := s.projectRepo.Store(ctx, entity.Project); err != nil {
		return nil, err
	}

	return &UpdateProjectOut{}, nil
}

func (s *service) RemoveProject(ctx context.Context, in *RemoveProjectIn) (*RemoveProjectOut, error) {
	entity, err := s.projectRepo.Find(ctx, in.ProjectID)
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, errors.New("Entity not found")
	}

	if err := s.projectRepo.Remove(ctx, in.ProjectID); err != nil {
		return nil, err
	}
	return &RemoveProjectOut{}, nil
}

type ViewProject struct {
	*project.Project
}
