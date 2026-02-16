package postsvc

import (
	"context"
	"errors"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/media"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/post"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/mediasvc"
	"github.com/gabriel-vasile/mimetype"
)

type PostRepository interface {
	Find(ctx context.Context, id post.PostID) (*ViewPost, error)
	Store(ctx context.Context, entity *post.Post) error
	Remove(ctx context.Context, ids ...post.PostID) error
}

type PostService interface {
	List(ctx context.Context, cond post.ListCond, option ListOption) ([]*ViewPost, error)
	Count(ctx context.Context, cond post.CountCond, option CountOption) (*uint64, error)
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
	ShowPosts(ctx context.Context, in *ShowPostsIn) (*ShowPostsOut, error)
	ShowPost(ctx context.Context, in *ShowPostIn) (*ShowPostOut, error)
	CreatePost(ctx context.Context, in *CreatePostIn) (*CreatePostOut, error)
	UpdatePost(ctx context.Context, in *UpdatePostIn) (*UpdatePostOut, error)
	RemovePost(ctx context.Context, in *RemovePostIn) (*RemovePostOut, error)
}

// Todo: pagination
type ShowPostsIn struct{}
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

type service struct {
	postRepo  PostRepository
	postSvc   PostService
	mediaRepo mediasvc.MediaRepository
	mediaSvc  mediasvc.MediaService
}

func NewService(postRepo PostRepository, postSvc PostService, mediaRepo mediasvc.MediaRepository, mediaSvc mediasvc.MediaService) Service {
	return &service{
		postRepo:  postRepo,
		postSvc:   postSvc,
		mediaRepo: mediaRepo,
		mediaSvc:  mediaSvc,
	}
}

func (s *service) ShowPosts(ctx context.Context, in *ShowPostsIn) (*ShowPostsOut, error) {
	cond := &post.ListCond{}
	opt := &ListOption{}
	entities, err := s.postSvc.List(ctx, *cond, *opt)
	if err != nil {
		return nil, err
	}

	for _, entity := range entities {
		postID := media.PostID(entity.PostID)
		mediaCond := &media.ListCond{
			MediaPostID: &postID,
		}

		mediaOpt := &mediasvc.ListOption{}
		mediaList, err := s.mediaSvc.List(ctx, *mediaCond, *mediaOpt)
		if err != nil {
			return nil, err
		}

		entity.MediaList = append(entity.MediaList, mediaList...)
	}

	countCond := &post.CountCond{}
	countOpt := &CountOption{}
	count, err := s.postSvc.Count(ctx, *countCond, *countOpt)
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

	postID := media.PostID(entity.PostID)
	mediaCond := &media.ListCond{
		MediaPostID: &postID,
	}

	mediaOpt := &mediasvc.ListOption{}
	mediaList, err := s.mediaSvc.List(ctx, *mediaCond, *mediaOpt)
	if err != nil {
		return nil, err
	}

	entity.MediaList = mediaList

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

type ViewPost struct {
	post.Post
	// linked other domain here whenever you need them

	MediaList []*mediasvc.ViewMedia
}

type MediaImage struct {
	Filename    string
	ContentType string
	Content     []byte
}
