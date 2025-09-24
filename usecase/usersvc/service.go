package usersvc

import (
	"context"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/user"
)

type UserRepository interface {
	Find(ctx context.Context, id string) (*user.User, error)
	Store(ctx context.Context, entity *user.User) error
	Remove(ctx context.Context, ids ...string) error
}

type UserService interface {
	List(ctx context.Context, cond user.ListCond, option ListOption) ([]*ViewUser, error)
	Count(ctx context.Context, cond user.CountCond, option CountOption) (*uint64, error)
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
	ShowUser(ctx context.Context, in *ShowUserIn) (*ShowUserOut, error)
}

type ShowUserIn struct{}

type ShowUserOut struct {
	Item ViewUser
}

type ViewUser struct {
	user.User
	// linked other domain here whenever you need them

}
