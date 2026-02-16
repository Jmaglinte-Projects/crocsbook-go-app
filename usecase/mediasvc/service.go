package mediasvc

import (
	"context"
	"errors"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/media"
	"github.com/gabriel-vasile/mimetype"
)

type MediaRepository interface {
	Find(ctx context.Context, id media.MediaID) (*ViewMedia, error)
	Store(ctx context.Context, entity *media.Media) error
	Remove(ctx context.Context, ids ...media.MediaID) error
}

type MediaR2Repository interface {
	// Find returns a presigned URL for the object at key (for frontend display).
	Find(ctx context.Context, key string) (string, error)
	Store(ctx context.Context, entity *media.Media) error
	Remove(ctx context.Context, keys ...string) error
}

type MediaService interface {
	List(ctx context.Context, cond media.ListCond, option ListOption) ([]*ViewMedia, error)
	Count(ctx context.Context, cond media.CountCond, option CountOption) (*uint64, error)
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
	ShowMedias(ctx context.Context, in *ShowMediasIn) (*ShowMediasOut, error)
	ShowMedia(ctx context.Context, in *ShowMediaIn) (*ShowMediaOut, error)
	CreateMedia(ctx context.Context, in *CreateMediaIn) (*CreateMediaOut, error)
	UpdateMedia(ctx context.Context, in *UpdateMediaIn) (*UpdateMediaOut, error)
	RemoveMedia(ctx context.Context, in *RemoveMediaIn) (*RemoveMediaOut, error)
}

type ShowMediasIn struct{}
type ShowMediasOut struct {
	Items []*ViewMedia
	Total *uint64
}

type ShowMediaIn struct {
	MediaID media.MediaID
}
type ShowMediaOut struct {
	Item *ViewMedia
}

type CreateMediaIn struct {
	MediaPostID media.PostID
	// Type        media.Type

	Data []byte
}
type CreateMediaOut struct{}

type UpdateMediaIn struct {
	MediaID media.MediaID

	// refactor this later
	URL *string
}
type UpdateMediaOut struct{}

type RemoveMediaIn struct {
	MediaID media.MediaID
}
type RemoveMediaOut struct{}

type service struct {
	mediaRepo MediaRepository
	mediaSvc  MediaService
}

func NewService(mediaRepo MediaRepository, mediaSvc MediaService) Service {
	return &service{
		mediaRepo: mediaRepo,
		mediaSvc:  mediaSvc,
	}
}

func (s *service) ShowMedias(ctx context.Context, in *ShowMediasIn) (*ShowMediasOut, error) {
	cond := &media.ListCond{}
	opt := &ListOption{}
	entities, err := s.mediaSvc.List(ctx, *cond, *opt)
	if err != nil {
		return nil, err
	}

	countCond := &media.CountCond{}
	countOpt := &CountOption{}
	count, err := s.mediaSvc.Count(ctx, *countCond, *countOpt)
	if err != nil {
		return nil, err
	}

	return &ShowMediasOut{
		Items: entities,
		Total: count,
	}, nil
}
func (s *service) ShowMedia(ctx context.Context, in *ShowMediaIn) (*ShowMediaOut, error) {
	entity, err := s.mediaRepo.Find(ctx, in.MediaID)
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, errors.New("Entity not found")
	}

	return &ShowMediaOut{
		Item: entity,
	}, nil
}

func (s *service) CreateMedia(ctx context.Context, in *CreateMediaIn) (*CreateMediaOut, error) {
	now := time.Now()

	id, err := media.NewMediaID()
	if err != nil {
		return nil, err
	}

	entity := &media.Media{}
	entity.MediaID = id
	entity.MediaPostID = in.MediaPostID

	// Refactor this later
	// entity.Type = in.Type
	mt := mimetype.Detect(in.Data)
	entity.MediaSet.Content = in.Data
	entity.MediaSet.ContentType = mt.String()
	entity.CreatedTime = now

	if err = s.mediaRepo.Store(ctx, entity); err != nil {
		return nil, err
	}

	return &CreateMediaOut{}, nil
}

func (s *service) UpdateMedia(ctx context.Context, in *UpdateMediaIn) (*UpdateMediaOut, error) {
	entity, err := s.mediaRepo.Find(ctx, in.MediaID)
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, errors.New("Entity not found")
	}

	// entity.URL = in.URL
	// todo Updated time | modify sql migration

	return &UpdateMediaOut{}, nil
}
func (s *service) RemoveMedia(ctx context.Context, in *RemoveMediaIn) (*RemoveMediaOut, error) {
	entity, err := s.mediaRepo.Find(ctx, in.MediaID)
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, errors.New("Entity not found")
	}

	if err = s.mediaRepo.Remove(ctx, in.MediaID); err != nil {
		return nil, err
	}

	return &RemoveMediaOut{}, nil
}

type ViewMedia struct {
	media.Media

	// URL is a presigned link to display the file (e.g. for img src).
	PresignedURL string
}
