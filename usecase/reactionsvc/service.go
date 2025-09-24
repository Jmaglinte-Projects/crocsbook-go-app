package reactionsvc

import (
	"context"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/reaction"
)

type ReactionRepository interface {
	Find(ctx context.Context, id string) (*reaction.Reaction, error)
	Store(ctx context.Context, entity *reaction.Reaction) error
	Remove(ctx context.Context, ids ...string) error
}

type ReactionService interface {
	List(ctx context.Context, cond reaction.ListCond, option ListOption) ([]*ViewReaction, error)
	Count(ctx context.Context, cond reaction.CountCond, option CountOption) (*uint64, error)
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
	ShowReaction(ctx context.Context, in *ShowReactionIn) (*ShowReactionOut, error)
}

type ShowReactionIn struct{}

type ShowReactionOut struct {
	Item ViewReaction
}

type ViewReaction struct {
	reaction.Reaction
	// linked other domain here whenever you need them

}
