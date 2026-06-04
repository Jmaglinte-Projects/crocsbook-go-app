package grpc

import (
	"context"

	pb "github.com/Jmaglinte-Projects/crocsbook-go-app/infra/grpc/lib"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/authsvc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authServer struct {
	pb.UnimplementedAuthServiceServer
	svc authsvc.Service
}

// NewAuthHandler returns a gRPC server that implements AuthService.
func NewAuthHandler(svc authsvc.Service) pb.AuthServiceServer {
	return &authServer{svc: svc}
}

func (s *authServer) GoogleSignIn(ctx context.Context, req *pb.GoogleSignInIn) (*pb.GoogleSignInOut, error) {
	in := &authsvc.GoogleSignInIn{IDToken: req.GetIdToken()}
	out, err := s.svc.GoogleSignIn(ctx, in)
	if err != nil {
		return nil, mapAuthErr(err)
	}
	return &pb.GoogleSignInOut{JwtToken: out.JwtToken}, nil
}

func (s *authServer) RegisterUser(ctx context.Context, req *pb.RegisterUserIn) (*pb.RegisterUserOut, error) {
	in := &authsvc.RegisterUserIn{
		Email:    req.GetEmail(),
		Gender:   req.GetGender(),
		Nickname: req.GetNickname(),
		Username: req.GetUsername(),
		Password: req.GetPassword(),
	}
	out, err := s.svc.RegisterUser(ctx, in)
	if err != nil {
		return nil, mapAuthErr(err)
	}
	return &pb.RegisterUserOut{JwtToken: out.JwtToken}, nil
}

func (s *authServer) SignIn(ctx context.Context, req *pb.SignInIn) (*pb.SignInOut, error) {
	in := &authsvc.SignInIn{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}
	out, err := s.svc.SignIn(ctx, in)
	if err != nil {
		return nil, mapAuthErr(err)
	}
	return &pb.SignInOut{JwtToken: out.JwtToken}, nil
}

func mapAuthErr(err error) error {
	switch err {
	case authsvc.ErrInvalidIDToken:
		return status.Error(codes.Unauthenticated, err.Error())
	case authsvc.ErrInvalidCredentials:
		return status.Error(codes.Unauthenticated, err.Error())
	case authsvc.ErrEmailAlreadyExists:
		return status.Error(codes.AlreadyExists, err.Error())
	case authsvc.ErrInvalidRegisterForm:
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return err
	}
}
