package grpc

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func strPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func strPtrToProto(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

func timePtrToProto(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}
