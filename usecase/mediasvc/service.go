package mediasvc

import (
	"context"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/media"
)

type MediaRepository interface {
	Find(ctx context.Context, id string) (*media.Media, error)
	Store(ctx context.Context, entity *media.Media) error
	Remove(ctx context.Context, ids ...string) error
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
	ShowMedia(ctx context.Context, in *ShowMediaIn) (*ShowMediaOut, error)
}

type ShowMediaIn struct{}

type ShowMediaOut struct {
	Item ViewMedia
}

type ViewMedia struct {
	media.Media
	// linked other domain here whenever you need them

}
