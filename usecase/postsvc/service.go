package postsvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/media"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/post"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/project"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/mediasvc"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/projectsvc"
	"github.com/gabriel-vasile/mimetype"
)

type PostRepository interface {
	Find(ctx context.Context, id post.PostID) (*ViewPost, error)
	Store(ctx context.Context, entity *post.Post) error
	Remove(ctx context.Context, ids ...post.PostID) error
}

type PostService interface {
	List(ctx context.Context, cond post.ListCond) ([]*ViewPost, error)
	Count(ctx context.Context, cond post.CountCond) (*uint64, error)
	ListPostStatsByProjectIds(ctx context.Context, cond post.ListPostStatsByProjectIdsCond, projectId ...post.ProjectID) ([]*ListPostStatsByProjectIds, error)
}

type Service interface {
	ShowPosts(ctx context.Context, in *ShowPostsIn) (*ShowPostsOut, error)
	ShowPost(ctx context.Context, in *ShowPostIn) (*ShowPostOut, error)
	CreatePost(ctx context.Context, in *CreatePostIn) (*CreatePostOut, error)
	UpdatePost(ctx context.Context, in *UpdatePostIn) (*UpdatePostOut, error)
	RemovePost(ctx context.Context, in *RemovePostIn) (*RemovePostOut, error)

	ShowPostByProjectId(ctx context.Context, in *ShowPostByProjectIdIn) (*ShowPostByProjectIdOut, error)
}

// Todo: pagination
type ShowPostsIn struct {
	Filter *Filter
}
type ShowPostsOut struct {
	Items []*ViewPost
	Total uint64
}

type ShowPostIn struct {
	PostID post.PostID
}
type ShowPostOut struct {
	Item *ViewPost
}

type CreatePostIn struct {
	PostProjectID post.ProjectID
	Content       *string
	Visibility    *post.Visibility

	MediaList [][]byte
}
type CreatePostOut struct{}

type UpdatePostIn struct {
	PostID        post.PostID
	PostProjectID post.ProjectID

	Content    *string
	Visibility *post.Visibility
}
type UpdatePostOut struct{}

type RemovePostIn struct {
	PostID post.PostID
}
type RemovePostOut struct{}

type ShowPostByProjectIdIn struct {
	ProjectID post.ProjectID

	Filter *Filter
}
type ShowPostByProjectIdOut struct {
	Items []*ViewPost
	Total uint64
}

type service struct {
	postRepo      PostRepository
	postSvc       PostService
	mediaRepo     mediasvc.MediaRepository
	mediaSvc      mediasvc.MediaService
	projectSvc    projectsvc.ProjectService
	projectR2Repo projectsvc.ProjectR2Repository
}

func NewService(postRepo PostRepository, postSvc PostService, mediaRepo mediasvc.MediaRepository, mediaSvc mediasvc.MediaService, projectSvc projectsvc.ProjectService, projectR2Repo projectsvc.ProjectR2Repository) Service {
	return &service{
		postRepo:      postRepo,
		postSvc:       postSvc,
		mediaRepo:     mediaRepo,
		mediaSvc:      mediaSvc,
		projectSvc:    projectSvc,
		projectR2Repo: projectR2Repo,
	}
}

func (s *service) ShowPosts(ctx context.Context, in *ShowPostsIn) (*ShowPostsOut, error) {
	cond := &post.ListCond{}
	in.Filter.Unmarshal(cond)

	entities, err := s.postSvc.List(ctx, *cond)
	if err != nil {
		return nil, err
	}

	if err := s.setProject(ctx, entities...); err != nil {
		return nil, err
	}

	if err := s.setMedia(ctx, entities...); err != nil {
		return nil, err
	}

	if err := s.setPostStatsByProjectIds(ctx, entities...); err != nil {
		return nil, err
	}

	// b, _ := json.MarshalIndent(entities, "", "  ")
	// fmt.Println("entities:", string(b))

	countCond := &post.CountCond{}
	count, err := s.postSvc.Count(ctx, *countCond)
	if err != nil {
		return nil, err
	}

	return &ShowPostsOut{
		Items: entities,
		Total: *count,
	}, nil
}

func (s *service) ShowPost(ctx context.Context, in *ShowPostIn) (*ShowPostOut, error) {
	entity, err := s.postRepo.Find(ctx, post.PostID(in.PostID))
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, errors.New("Entity not found")
	}

	if err := s.setMedia(ctx, entity); err != nil {
		return nil, err
	}

	if err := s.setProject(ctx, entity); err != nil {
		return nil, err
	}

	b, _ := json.MarshalIndent(entity.Project, "", "  ")
	fmt.Println("entity.Project:", string(b))

	return &ShowPostOut{
		Item: entity,
	}, nil
}

func (s *service) CreatePost(ctx context.Context, in *CreatePostIn) (*CreatePostOut, error) {
	now := time.Now()

	id, err := post.NewPostID()
	if err != nil {
		return nil, err
	}

	entity := &post.Post{}
	entity.PostID = id
	entity.PostProjectID = in.PostProjectID
	entity.Content = in.Content
	entity.Visibility = in.Visibility
	entity.CreatedTime = now

	err = s.postRepo.Store(ctx, entity)
	if err != nil {
		return nil, err
	}

	for _, mediaItem := range in.MediaList {
		createMediaIn := &mediasvc.CreateMediaIn{
			MediaPostID: media.PostID(entity.PostID),
			Data:        mediaItem,
		}
		_, err = s.createMedia(ctx, createMediaIn)
		if err != nil {
			return nil, err
		}
	}

	return &CreatePostOut{}, nil
}

func (s *service) UpdatePost(ctx context.Context, in *UpdatePostIn) (*UpdatePostOut, error) {
	now := time.Now()

	entity, err := s.postRepo.Find(ctx, in.PostID)
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, errors.New("Entity not found")
	}

	entity.Post.Content = in.Content
	entity.Visibility = in.Visibility
	entity.UpdatedTime = &now

	err = s.postRepo.Store(ctx, &entity.Post)
	if err != nil {
		return nil, err
	}

	// TODO: update media feature

	return &UpdatePostOut{}, nil
}

func (s *service) RemovePost(ctx context.Context, in *RemovePostIn) (*RemovePostOut, error) {
	entity, err := s.postRepo.Find(ctx, in.PostID)
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, errors.New("Entity not found")
	}

	err = s.postRepo.Remove(ctx, entity.PostID)
	if err != nil {
		return nil, err
	}

	return &RemovePostOut{}, nil
}

func (s *service) ShowPostByProjectId(ctx context.Context, in *ShowPostByProjectIdIn) (*ShowPostByProjectIdOut, error) {
	cond := &post.ListCond{
		PostProjectID: &in.ProjectID,
	}
	// for filters
	in.Filter.Unmarshal(cond)

	b, _ := json.MarshalIndent(cond, "", "  ")
	fmt.Println("ShowPostByProjectId cond:", string(b))

	entities, err := s.postSvc.List(ctx, *cond)
	if err != nil {
		return nil, err
	}

	if err := s.setMedia(ctx, entities...); err != nil {
		return nil, err
	}

	if err := s.setProject(ctx, entities...); err != nil {
		return nil, err
	}

	return &ShowPostByProjectIdOut{
		Items: entities,
		Total: uint64(len(entities)),
	}, nil
}

func (s *service) createMedia(ctx context.Context, in *mediasvc.CreateMediaIn) (*mediasvc.CreateMediaOut, error) {
	now := time.Now()

	id, err := media.NewMediaID()
	if err != nil {
		return nil, err
	}

	entity := &media.Media{}
	entity.MediaID = id
	entity.MediaPostID = in.MediaPostID

	// entity.Type = in.Type
	mt := mimetype.Detect(in.Data)
	entity.MediaSet.Content = in.Data
	entity.MediaSet.ContentType = mt.String()
	entity.CreatedTime = now

	if err = s.mediaRepo.Store(ctx, entity); err != nil {
		return nil, err
	}

	return &mediasvc.CreateMediaOut{}, nil
}

func (s *service) setMedia(ctx context.Context, entities ...*ViewPost) error {
	for _, entity := range entities {
		postID := media.PostID(entity.PostID)
		mediaCond := &media.ListCond{
			MediaPostID: &postID,
		}

		mediaOpt := &mediasvc.ListOption{}
		mediaList, err := s.mediaSvc.List(ctx, *mediaCond, *mediaOpt)
		if err != nil {
			fmt.Println("Error listing media")
			return err
		}

		entity.MediaList = append(entity.MediaList, mediaList...)
	}
	return nil
}

func (s *service) setProject(ctx context.Context, entities ...*ViewPost) error {
	projectIDs := []project.ProjectID{}
	for _, entity := range entities {
		projectIDs = append(projectIDs, project.ProjectID(entity.PostProjectID))
	}

	projectCond := &project.ListCond{
		ProjectIDs: projectIDs,
	}
	projectOpt := &projectsvc.ListOption{}
	projects, err := s.projectSvc.List(ctx, *projectCond, *projectOpt)
	if err != nil {
		fmt.Println("Error listing projects")
		return err
	}

	projectByID := make(map[project.ProjectID]*projectsvc.ViewProject)
	for _, p := range projects {
		projectByID[p.ProjectID] = p
	}
	for _, entity := range entities {
		if p, ok := projectByID[project.ProjectID(entity.PostProjectID)]; ok {
			presignedURL := ""
			if p.ThumbnailKey != nil {
				if u, err := s.projectR2Repo.Find(ctx, *p.ThumbnailKey); err == nil {
					presignedURL = u
				}
			}
			entity.Project = &ViewProject{
				Project:      p.Project,
				ThumbnailURL: presignedURL,
			}
		}
	}
	return nil
}

func (s *service) setPostStatsByProjectIds(ctx context.Context, entities ...*ViewPost) error {
	projectIDs := []post.ProjectID{}
	for _, entity := range entities {
		projectIDs = append(projectIDs, post.ProjectID(entity.PostProjectID))
	}

	cond := &post.ListPostStatsByProjectIdsCond{}
	out, err := s.postSvc.ListPostStatsByProjectIds(ctx, *cond, projectIDs...)
	if err != nil {
		fmt.Println("Error counting total posts by project ID")
		return err
	}

	projectByID := make(map[project.ProjectID]*ListPostStatsByProjectIds)
	for _, p := range out {
		projectByID[project.ProjectID(p.ProjectID)] = p
	}

	for _, entity := range entities {
		if p, ok := projectByID[project.ProjectID(entity.PostProjectID)]; ok {
			entity.PostCount = p.Count
			entity.LastPostTime = p.LastPostTime
		}
	}

	return nil
}

type Filter struct {
	CreatedTime *time.Time

	SortKey post.PostSortKey
	Size    int64
	Offset  *int64
}

func (src *Filter) Unmarshal(dest *post.ListCond) {
	if src.CreatedTime != nil {
		dest.CreatedTime = src.CreatedTime
	}

	dest.SortKey = src.SortKey
	dest.Size = src.Size

	if src.Offset != nil {
		dest.Offset = src.Offset
	}
}

type ViewPost struct {
	post.Post

	// linked other domain here whenever you need them
	MediaList []*mediasvc.ViewMedia // it should be under postsvc not on mediasvc refactor this later
	Project   *ViewProject

	PostCount    uint64
	LastPostTime time.Time
}

type MediaImage struct {
	Filename    string
	ContentType string
	Content     []byte
}

type ViewProject struct {
	project.Project

	ThumbnailURL string
}

type ListPostStatsByProjectIds struct {
	ProjectID    post.ProjectID
	Count        uint64
	LastPostTime time.Time
}
