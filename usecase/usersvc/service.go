package usersvc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/user"
	"github.com/gabriel-vasile/mimetype"
)

type UserRepository interface {
	Find(ctx context.Context, id user.UserID) (*ViewUser, error)
	Store(ctx context.Context, entity *user.User) error
	Remove(ctx context.Context, ids ...user.UserID) error
}

type UserR2Repository interface {
	// Find returns a presigned URL for the object at key (for frontend display).
	// TODO: use public domain URL instead of presigned URL
	Find(ctx context.Context, key string) (string, error)
	Store(ctx context.Context, entity *user.User) error
	Remove(ctx context.Context, keys ...string) error
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
	ShowUsers(ctx context.Context, in *ShowUsersIn) (*ShowUsersOut, error)
	ShowUser(ctx context.Context, in *ShowUserIn) (*ShowUserOut, error)
	CreateUser(ctx context.Context, in *CreateUserIn) (*CreateUserOut, error)
	UpdateUser(ctx context.Context, in *UpdateUserIn) (*UpdateUserOut, error)
	RemoveUser(ctx context.Context, in *RemoveUserIn) (*RemoveUserOut, error)
}

type ShowUsersIn struct{}
type ShowUsersOut struct {
	Items []*ViewUser
	Total uint64
}

type ShowUserIn struct {
	UserID user.UserID
}
type ShowUserOut struct {
	Item *ViewUser
}

type CreateUserIn struct {
	Email    string
	Gender   user.Gender
	Nickname *string
	Username *string
	Password *string
}
type CreateUserOut struct{}

type UpdateUserIn struct {
	UserID user.UserID

	Email    string
	Password *string
	Gender   user.Gender
	Nickname *string

	ImageData []byte
}
type UpdateUserOut struct{}

type RemoveUserIn struct {
	UserID user.UserID
}
type RemoveUserOut struct{}

type service struct {
	userRepo   UserRepository
	userSvc    UserService
	userR2Repo UserR2Repository
}

func NewService(userRepo UserRepository, userSvc UserService, userR2Repo UserR2Repository) Service {
	return &service{
		userRepo:   userRepo,
		userSvc:    userSvc,
		userR2Repo: userR2Repo,
	}
}

func (s *service) ShowUsers(ctx context.Context, in *ShowUsersIn) (*ShowUsersOut, error) {
	cond := &user.ListCond{}
	opt := &ListOption{}
	entities, err := s.userSvc.List(ctx, *cond, *opt)
	if err != nil {
		return nil, err
	}

	countCond := &user.CountCond{}
	countOpt := &CountOption{}
	total, err := s.userSvc.Count(ctx, *countCond, *countOpt)
	if err != nil {
		return nil, err
	}

	return &ShowUsersOut{
		Items: entities,
		Total: *total,
	}, nil
}

func (s *service) ShowUser(ctx context.Context, in *ShowUserIn) (*ShowUserOut, error) {
	entity, err := s.userRepo.Find(ctx, in.UserID)
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, errors.New("User not found")
	}

	s.setProfileUrl(entity)

	return &ShowUserOut{Item: entity}, nil
}

func (s *service) CreateUser(ctx context.Context, in *CreateUserIn) (*CreateUserOut, error) {
	now := time.Now()
	id, err := user.NewUserID()
	if err != nil {
		return nil, err
	}

	entity := &user.User{}
	entity.UserID = id
	entity.Email = in.Email
	entity.Gender = in.Gender
	entity.Nickname = in.Nickname
	entity.Username = in.Username
	entity.Password = in.Password
	entity.CreatedTime = now

	err = s.userRepo.Store(ctx, entity)
	if err != nil {
		return nil, err
	}

	return &CreateUserOut{}, nil
}

func (s *service) UpdateUser(ctx context.Context, in *UpdateUserIn) (*UpdateUserOut, error) {
	now := time.Now()

	entity, err := s.userRepo.Find(ctx, in.UserID)
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, errors.New("Entity not found")
	}

	entity.Email = in.Email
	entity.Gender = in.Gender
	if in.Nickname != nil {
		entity.Nickname = in.Nickname
	}
	if in.Password != nil {
		entity.Password = in.Password
	}
	entity.UpdatedTime = &now

	if len(in.ImageData) > 0 {
		mt := mimetype.Detect(in.ImageData)
		imageSet := &user.ImageSet{
			ContentType: mt.String(),
			Content:     in.ImageData,
		}
		entity.ImageSet = imageSet

		if err = s.userR2Repo.Store(ctx, &entity.User); err != nil {
			fmt.Println("Error storing image to r2")
			return nil, err
		}
	}

	err = s.userRepo.Store(ctx, &entity.User)
	if err != nil {
		return nil, err
	}
	return &UpdateUserOut{}, nil
}
func (s *service) RemoveUser(ctx context.Context, in *RemoveUserIn) (*RemoveUserOut, error) {
	err := s.userRepo.Remove(ctx, in.UserID)
	if err != nil {
		return nil, err
	}

	return &RemoveUserOut{}, nil
}

func (s *service) setProfileUrl(entity *ViewUser) {
	if entity.User.ProfileKey != nil {
		url := fmt.Sprintf("%s/%s", os.Getenv("R2_PUBLIC_BASE_URL"), *entity.User.ProfileKey)
		entity.User.ProfileURL = &url
	}
}

type ViewUser struct {
	user.User
	// linked other domain here whenever you need them
}
