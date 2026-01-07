package reactionsvc

import (
	"context"
	"errors"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/reaction"
)

type ReactionRepository interface {
	Find(ctx context.Context, id reaction.ReactionID) (*reaction.Reaction, error)
	Store(ctx context.Context, entity *reaction.Reaction) error
	Remove(ctx context.Context, ids ...reaction.ReactionID) error
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
	ShowReactions(ctx context.Context, in *ShowReactionsIn) (*ShowReactionsOut, error)
	ShowReaction(ctx context.Context, in *ShowReactionIn) (*ShowReactionOut, error)
	CreateReaction(ctx context.Context, in *CreateReactionIn) (*CreateReactionOut, error)
	UpdateReaction(ctx context.Context, in *UpdateReactionIn) (*UpdateReactionOut, error)
	RemoveReaction(ctx context.Context, in *RemoveReactionIn) (*RemoveReactionOut, error)
}

type ShowReactionsIn struct{}
type ShowReactionsOut struct {
	Items []*ViewReaction
	Total uint64
}

type ShowReactionIn struct {
	ReactionID reaction.ReactionID
}
type ShowReactionOut struct {
	Item reaction.Reaction
}
type CreateReactionIn struct {
	ReactionProjectID reaction.ProjectID
}
type CreateReactionOut struct{}

type UpdateReactionIn struct {
	ReactionID        reaction.ReactionID
	ReactionProjectID reaction.ProjectID
}
type UpdateReactionOut struct{}

type RemoveReactionIn struct {
	ReactionID reaction.ReactionID
}
type RemoveReactionOut struct{}

type service struct {
	reactionRepo ReactionRepository
	reactionSvc  ReactionService
}

type ServiceDep struct {
	ReactionRepo ReactionRepository
	ReactionSvc  ReactionService
}

func NewService(dep ServiceDep) Service {
	return &service{
		reactionRepo: dep.ReactionRepo,
		reactionSvc:  dep.ReactionSvc,
	}
}

func (s *service) ShowReactions(ctx context.Context, in *ShowReactionsIn) (*ShowReactionsOut, error) {
	cond := &reaction.ListCond{}
	opt := &ListOption{}
	entities, err := s.reactionSvc.List(ctx, *cond, *opt)
	if err != nil {
		return nil, err
	}

	if entities == nil {
		return nil, errors.New("entities not found")
	}

	countCond := &reaction.CountCond{}
	countOpt := &CountOption{}
	count, err := s.reactionSvc.Count(ctx, *countCond, *countOpt)
	if err != nil {
		return nil, err
	}

	return &ShowReactionsOut{
		Items: entities,
		Total: *count,
	}, nil
}

func (s *service) ShowReaction(ctx context.Context, in *ShowReactionIn) (*ShowReactionOut, error) {
	entity, err := s.reactionRepo.Find(ctx, in.ReactionID)
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, errors.New("entity not found")
	}

	return &ShowReactionOut{Item: *entity}, nil
}

func (s *service) CreateReaction(ctx context.Context, in *CreateReactionIn) (*CreateReactionOut, error) {
	now := time.Now()
	id, err := reaction.NewReactionID()
	if err != nil {
		return nil, err
	}

	entity := &reaction.Reaction{}
	entity.ReactionID = id
	entity.ReactionProjectID = in.ReactionProjectID

	like := reaction.Type_Like
	entity.Type = &like
	entity.CreatedTime = now

	err = s.reactionRepo.Store(ctx, entity)
	if err != nil {
		return nil, err
	}

	return &CreateReactionOut{}, nil
}

func (s *service) UpdateReaction(ctx context.Context, in *UpdateReactionIn) (*UpdateReactionOut, error) {
	now := time.Now()

	entity, err := s.reactionRepo.Find(ctx, in.ReactionID)
	if err != nil {
		return nil, err
	}

	if entity != nil {
		err := s.reactionRepo.Remove(ctx, entity.ReactionID)
		if err != nil {
			return nil, err
		}
	} else {
		id, err := reaction.NewReactionID()
		if err != nil {
			return nil, err
		}

		entity := &reaction.Reaction{}
		entity.ReactionID = id
		entity.ReactionProjectID = in.ReactionProjectID

		like := reaction.Type_Like
		entity.Type = &like
		entity.CreatedTime = now

		err = s.reactionRepo.Store(ctx, entity)
		if err != nil {
			return nil, err
		}
	}

	return &UpdateReactionOut{}, nil
}

func (s *service) RemoveReaction(ctx context.Context, in *RemoveReactionIn) (*RemoveReactionOut, error) {
	if err := s.reactionRepo.Remove(ctx, in.ReactionID); err != nil {
		return nil, err
	}

	return &RemoveReactionOut{}, nil
}

type ViewReaction struct {
	reaction.Reaction
	// linked other domain here whenever you need them

}
