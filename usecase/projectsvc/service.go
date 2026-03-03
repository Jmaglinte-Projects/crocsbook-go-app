package projectsvc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/project"
	"github.com/gabriel-vasile/mimetype"
)

type ProjectRepository interface {
	Find(ctx context.Context, id project.ProjectID) (*ViewProject, error)
	Store(ctx context.Context, entity *project.Project) error
	Remove(ctx context.Context, ids ...project.ProjectID) error
}

type ProjectLikeRepository interface {
	Find(ctx context.Context, projectID project.ProjectID, userID project.UserID) (*project.ProjectLike, error)
	Store(ctx context.Context, entity *project.ProjectLike) error
	Remove(ctx context.Context, ids ...RemoveCond) error
}

type RemoveCond struct {
	ProjectID project.ProjectID
	UserID    project.UserID
}

type ProjectR2Repository interface {
	// Find returns a presigned URL for the object at key (for frontend display).
	Find(ctx context.Context, key string) (string, error)
	Store(ctx context.Context, entity *project.Project) error
	Remove(ctx context.Context, keys ...string) error
}

type ProjectService interface {
	List(ctx context.Context, cond project.ListCond, option ListOption) ([]*ViewProject, error)
	Count(ctx context.Context, cond project.CountCond, option CountOption) (*uint64, error)
}

type ProjectLikeService interface {
	List(ctx context.Context, cond project.ProjectLikeListCond) ([]*project.ProjectLike, error)
	Count(ctx context.Context, cond project.ProjectLikeCountCond) (*uint64, error)
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
	CountProjectLikes(ctx context.Context, in *CountProjectLikesIn) (*CountProjectLikesOut, error)
	LikeProject(ctx context.Context, in *LikeProjectIn) (*LikeProjectOut, error)
	UnlikeProject(ctx context.Context, in *UnlikeProjectIn) (*UnlikeProjectOut, error)
	CheckProjectLiked(ctx context.Context, in *CheckProjectLikedIn) (*CheckProjectLikedOut, error)
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
	Location       *string
	Cost           *int64
	StartDate      *time.Time
	CompletionDate *time.Time

	ThumbnailContent []byte
}
type CreateProjectOut struct{}

type UpdateProjectIn struct {
	ProjectID project.ProjectID

	Name           string
	Description    *string
	Location       *string
	Cost           *int64
	StartDate      *time.Time
	CompletionDate *time.Time

	ThumbnailContent []byte
}
type UpdateProjectOut struct{}

type RemoveProjectIn struct {
	ProjectID project.ProjectID
}
type RemoveProjectOut struct{}

type CountProjectLikesIn struct {
	ProjectID project.ProjectID
}
type CountProjectLikesOut struct {
	Count uint64
}
type LikeProjectIn struct {
	ProjectID project.ProjectID
	UserID    project.UserID
}
type LikeProjectOut struct{}
type UnlikeProjectIn struct {
	ProjectID project.ProjectID
	UserID    project.UserID
}
type UnlikeProjectOut struct{}

type CheckProjectLikedIn struct {
	ProjectID project.ProjectID
	UserID    project.UserID
}
type CheckProjectLikedOut struct {
	Liked bool
}

type service struct {
	projectRepo     ProjectRepository
	projectSvc      ProjectService
	projectR2Repo   ProjectR2Repository
	projectLikeRepo ProjectLikeRepository
	projectLikeSvc  ProjectLikeService
}

func NewService(projectRepo ProjectRepository, projectSvc ProjectService, projectR2Repo ProjectR2Repository, projectLikeRepo ProjectLikeRepository, projectLikeSvc ProjectLikeService) Service {
	return &service{
		projectRepo:     projectRepo,
		projectSvc:      projectSvc,
		projectR2Repo:   projectR2Repo,
		projectLikeRepo: projectLikeRepo,
		projectLikeSvc:  projectLikeSvc,
	}
}

func (s *service) ShowProjects(ctx context.Context, in *ShowProjectsIn) (*ShowProjectsOut, error) {
	cond := &project.ListCond{}
	opt := &ListOption{}

	entities, err := s.projectSvc.List(ctx, *cond, *opt)
	if err != nil {
		return nil, err
	}

	for _, entity := range entities {
		if entity.ThumbnailKey == nil {
			continue
		}

		presignedURL, err := s.projectR2Repo.Find(ctx, *entity.ThumbnailKey)
		if err != nil {
			return nil, err
		}

		entity.ThumbnailURL = presignedURL
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

	presignedURL, err := s.projectR2Repo.Find(ctx, *entity.ThumbnailKey)
	if err != nil {
		return nil, err
	}
	entity.ThumbnailURL = presignedURL

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
	entity.Location = in.Location
	entity.Cost = in.Cost
	entity.StartDate = in.StartDate
	entity.CompletionDate = in.CompletionDate
	entity.CreatedTime = now

	mt := mimetype.Detect(in.ThumbnailContent)
	thumbnailSet := project.ThumbnailSet{
		ContentType: mt.String(),
		Content:     in.ThumbnailContent,
	}
	entity.ThumbnailSet = thumbnailSet

	if err = s.projectR2Repo.Store(ctx, entity); err != nil {
		fmt.Println("Error storing thumbnail to r2")
		return nil, err
	}

	if err = s.projectRepo.Store(ctx, entity); err != nil {
		fmt.Println("Error storing project to mysql")
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

	entity.Location = in.Location
	entity.Cost = in.Cost
	entity.StartDate = in.StartDate
	entity.CompletionDate = in.CompletionDate
	entity.UpdatedTime = &now

	if len(in.ThumbnailContent) > 0 {
		mt := mimetype.Detect(in.ThumbnailContent)
		thumbnailSet := project.ThumbnailSet{
			ContentType: mt.String(),
			Content:     in.ThumbnailContent,
		}
		entity.ThumbnailSet = thumbnailSet

		// thumbnail should be set after this
		if err = s.projectR2Repo.Store(ctx, &entity.Project); err != nil {
			fmt.Println("Error storing thumbnail to r2")
			return nil, err
		}
	}

	if err := s.projectRepo.Store(ctx, &entity.Project); err != nil {
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

	if entity.ThumbnailKey != nil {
		key := *entity.ThumbnailKey
		if err := s.projectR2Repo.Remove(ctx, key); err != nil {
			return nil, err
		}
	}

	return &RemoveProjectOut{}, nil
}

func (s *service) CountProjectLikes(ctx context.Context, in *CountProjectLikesIn) (*CountProjectLikesOut, error) {
	cond := &project.ProjectLikeCountCond{
		ProjectID: &in.ProjectID,
	}
	count, err := s.projectLikeSvc.Count(ctx, *cond)
	if err != nil {
		return nil, err
	}

	return &CountProjectLikesOut{
		Count: *count,
	}, nil
}

func (s *service) LikeProject(ctx context.Context, in *LikeProjectIn) (*LikeProjectOut, error) {
	fmt.Println("--------------------------------")
	fmt.Println("in.ProjectID: ", string(in.ProjectID))
	fmt.Println("in.UserID: ", string(in.UserID))
	fmt.Println("--------------------------------")

	likeEntity, err := s.projectLikeRepo.Find(ctx, in.ProjectID, in.UserID)
	if err != nil {
		return nil, err
	}

	if likeEntity != nil {
		fmt.Println("Project already liked")
		return nil, err
	}

	entity := &project.ProjectLike{}
	entity.ProjectID = in.ProjectID
	entity.UserID = in.UserID
	entity.CreatedTime = time.Now()

	if err := s.projectLikeRepo.Store(ctx, entity); err != nil {
		return nil, err
	}

	return &LikeProjectOut{}, nil
}

func (s *service) UnlikeProject(ctx context.Context, in *UnlikeProjectIn) (*UnlikeProjectOut, error) {
	likeEntity, err := s.projectLikeRepo.Find(ctx, in.ProjectID, in.UserID)
	if err != nil {
		return nil, err
	}

	if likeEntity == nil {
		fmt.Println("Project not liked")
		return nil, err
	}

	cond := RemoveCond{ProjectID: in.ProjectID, UserID: in.UserID}
	if err := s.projectLikeRepo.Remove(ctx, cond); err != nil {
		return nil, err
	}

	return &UnlikeProjectOut{}, nil
}

func (s *service) CheckProjectLiked(ctx context.Context, in *CheckProjectLikedIn) (*CheckProjectLikedOut, error) {
	likeEntity, err := s.projectLikeRepo.Find(ctx, in.ProjectID, in.UserID)
	if err != nil {
		return nil, err
	}

	if likeEntity == nil {
		return &CheckProjectLikedOut{Liked: false}, nil
	}

	return &CheckProjectLikedOut{Liked: true}, nil
}

type ViewProject struct {
	project.Project

	ThumbnailURL string
}
