package postsvc

import (
	"context"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/post"
)

type PostRepository interface {
	Find(ctx context.Context, id string) (*post.Post, error)
	Store(ctx context.Context, entity *post.Post) error
	Remove(ctx context.Context, ids ...string) error
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
	ShowPost(ctx context.Context, in *ShowPostIn) (*ShowPostOut, error)
}

type ShowPostIn struct{}

type ShowPostOut struct {
	Item ViewPost
}

type ViewPost struct {
	post.Post
	// linked other domain here whenever you need them

}
