package grpc

import (
	"context"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/user"
	pb "github.com/Jmaglinte-Projects/crocsbook-go-app/infra/grpc/lib"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/usersvc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type userServer struct {
	pb.UnimplementedUserServiceServer
	svc usersvc.Service
}

func NewUserHandler(svc usersvc.Service) pb.UserServiceServer {
	return &userServer{
		svc: svc,
	}
}

func (s *userServer) ShowUsers(ctx context.Context, req *pb.ShowUsersIn) (*pb.ShowUsersOut, error) {
	in := &usersvc.ShowUsersIn{}

	users, err := s.svc.ShowUsers(ctx, in)
	if err != nil {
		return nil, err
	}

	out := &pb.ShowUsersOut{
		Items: make([]*pb.ViewUser, len(users.Items)),
	}

	for _, project := range users.Items {
		item := &ViewUser{}
		item.UnmarshalOriginal(project)
		out.Items = append(out.Items, &item.ViewUser)
	}

	return out, nil
}

func (s *userServer) ShowUser(ctx context.Context, req *pb.ShowUserIn) (*pb.ShowUserOut, error) {
	in := &usersvc.ShowUserIn{
		UserID: user.UserID(req.GetUserId()),
	}

	user, err := s.svc.ShowUser(ctx, in)
	if err != nil {
		return nil, err
	}

	out := &pb.ShowUserOut{
		Item: &pb.ViewUser{},
	}

	item := &ViewUser{}
	item.UnmarshalOriginal(user.Item)
	out.Item = &item.ViewUser

	return out, nil
}

func (s *userServer) CreateUser(ctx context.Context, req *pb.CreateUserIn) (*pb.CreateUserOut, error) {
	in := &usersvc.CreateUserIn{}

	_, err := s.svc.CreateUser(ctx, in)
	if err != nil {
		return nil, err
	}

	return &pb.CreateUserOut{}, nil
}

func (s *userServer) UpdateUser(ctx context.Context, req *pb.UpdateUserIn) (*pb.UpdateUserOut, error) {
	password := req.GetPassword()
	nickname := req.GetNickname()
	in := &usersvc.UpdateUserIn{
		UserID:    user.UserID(req.GetUserId()),
		Email:     req.GetEmail(),
		Password:  &password,
		Gender:    userGenderToEntity(req.GetGender()),
		Nickname:  &nickname,
		ImageData: req.GetImageData(),
	}

	_, err := s.svc.UpdateUser(ctx, in)
	if err != nil {
		return nil, err
	}
	return &pb.UpdateUserOut{}, nil
}

func (s *userServer) RemoveUser(ctx context.Context, req *pb.RemoveUserIn) (*pb.RemoveUserOut, error) {
	in := &usersvc.RemoveUserIn{}

	_, err := s.svc.RemoveUser(ctx, in)
	if err != nil {
		return nil, err
	}
	return &pb.RemoveUserOut{}, nil
}

type ViewUser struct {
	pb.ViewUser
}

func (dest *ViewUser) UnmarshalOriginal(src *usersvc.ViewUser) {
	if dest.User == nil {
		dest.User = &pb.User{}
	}
	d := dest.User

	gender := userGenderToProto(src.Gender)

	d.UserId = string(src.UserID)
	d.Email = src.Email
	d.Gender = gender
	d.ProfileUrl = strPtr(src.ProfileURL)
	d.Nickname = strPtr(src.Nickname)
	d.Username = strPtr(src.Username)

	d.CreatedTime = timestamppb.New(src.CreatedTime)
	d.UpdatedTime = timePtrToProto(src.UpdatedTime)

}

func userGenderToProto(g user.Gender) pb.Gender {
	switch g {
	case user.Gender_Male:
		return pb.Gender_GENDER_MALE
	case user.Gender_Female:
		return pb.Gender_GENDER_FEMALE
	default:
		return pb.Gender_GENDER_UNSPECIFIED
	}
}

func userGenderToEntity(g pb.Gender) user.Gender {
	switch g {
	case pb.Gender_GENDER_MALE:
		return user.Gender_Male
	case pb.Gender_GENDER_FEMALE:
		return user.Gender_Female
	default:
		return ""
	}
}
