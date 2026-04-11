package grpc

import (
	"context"

	pb "github.com/Jmaglinte-Projects/crocsbook-go-app/infra/grpc/lib"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/authsvc"
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
		return nil, err
	}
	return &pb.GoogleSignInOut{JwtToken: out.JwtToken}, nil
}
