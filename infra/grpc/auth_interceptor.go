package grpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/user"
	pb "github.com/Jmaglinte-Projects/crocsbook-go-app/infra/grpc/lib"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/authsvc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ctxKeyUserID struct{}

// UserIDFromContext returns the authenticated user id if the request passed UnaryAuthInterceptor.
func UserIDFromContext(ctx context.Context) (user.UserID, bool) {
	id, ok := ctx.Value(ctxKeyUserID{}).(user.UserID)
	return id, ok
}

// UnaryAuthInterceptor validates `Authorization: Bearer <jwt>` on incoming metadata for all unary
// RPCs except public methods (e.g. GoogleSignIn). jwtSecret must match NewService.
func UnaryAuthInterceptor(jwtSecret string) grpc.UnaryServerInterceptor {
	public := map[string]struct{}{
		pb.AuthService_GoogleSignIn_FullMethodName: {},
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if _, ok := public[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		raw, err := bearerFromIncomingMetadata(ctx)
		if err != nil {
			return nil, err
		}

		userID, err := authsvc.ParseUserIDFromJWT(jwtSecret, raw)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		fmt.Println("UnaryAuthInterceptor userID: ", userID)
		ctx = context.WithValue(ctx, ctxKeyUserID{}, userID)
		return handler(ctx, req)
	}
}

func bearerFromIncomingMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "no metadata")
	}
	for _, v := range md.Get("authorization") {
		v = strings.TrimSpace(v)
		if len(v) > 7 && strings.EqualFold(v[:7], "bearer ") {
			t := strings.TrimSpace(v[7:])
			if t != "" {
				return t, nil
			}
		}
	}
	return "", status.Error(codes.Unauthenticated, "no bearer token")
}
