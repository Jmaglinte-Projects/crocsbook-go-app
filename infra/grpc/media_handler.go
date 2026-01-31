package grpc

import (
	"context"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/media"
	pb "github.com/Jmaglinte-Projects/crocsbook-go-app/infra/grpc/lib"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/mediasvc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mediaServer struct {
	pb.UnimplementedMediaServiceServer
	svc mediasvc.Service
}

func NewMediaHandler(svc mediasvc.Service) pb.MediaServiceServer {
	return &mediaServer{
		svc: svc,
	}
}

func (s *mediaServer) ShowMedias(ctx context.Context, req *pb.ShowMediasIn) (*pb.ShowMediasOut, error) {
	in := &mediasvc.ShowMediasIn{}
	medias, err := s.svc.ShowMedias(ctx, in)
	if err != nil {
		return nil, err
	}

	out := pb.ShowMediasOut{
		Items: make([]*pb.ViewMedia, len(medias.Items)),
	}
	for _, media := range medias.Items {
		item := &ViewMedia{}
		item.UnmarshalOriginal(media)
		out.Items = append(out.Items, &item.ViewMedia)
	}

	return &out, nil
}

func (s *mediaServer) ShowMedia(ctx context.Context, req *pb.ShowMediaIn) (*pb.ShowMediaOut, error) {
	in := &mediasvc.ShowMediaIn{
		MediaID: media.MediaID(req.MediaId),
	}
	media, err := s.svc.ShowMedia(ctx, in)
	if err != nil {
		return nil, err
	}

	out := pb.ShowMediaOut{
		Item: &pb.ViewMedia{},
	}
	item := &ViewMedia{}
	item.UnmarshalOriginal(media.Item)
	out.Item = &item.ViewMedia

	return &out, nil
}

func (s *mediaServer) CreateMedia(ctx context.Context, req *pb.CreateMediaIn) (*pb.CreateMediaOut, error) {
	mediaType := metaTypeProtoToEntity(req.Type)
	in := &mediasvc.CreateMediaIn{
		MediaProjectID: media.ProjectID(req.MediaProjectId),
		URL:            &req.Url,
		Type:           &mediaType,
	}

	_, err := s.svc.CreateMedia(ctx, in)
	if err != nil {
		return nil, err
	}

	return &pb.CreateMediaOut{}, nil
}

func (s *mediaServer) UpdateMedia(ctx context.Context, req *pb.UpdateMediaIn) (*pb.UpdateMediaOut, error) {
	in := &mediasvc.UpdateMediaIn{
		MediaID: media.MediaID(req.MediaId),
		URL:     &req.Url,
	}
	_, err := s.svc.UpdateMedia(ctx, in)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateMediaOut{}, nil
}

func (s *mediaServer) RemoveMedia(ctx context.Context, req *pb.RemoveMediaIn) (*pb.RemoveMediaOut, error) {
	in := &mediasvc.RemoveMediaIn{
		MediaID: media.MediaID(req.MediaId),
	}
	_, err := s.svc.RemoveMedia(ctx, in)
	if err != nil {
		return nil, err
	}

	return &pb.RemoveMediaOut{}, nil
}

type ViewMedia struct {
	pb.ViewMedia
}

func (dest *ViewMedia) UnmarshalOriginal(src *mediasvc.ViewMedia) {
	if dest.Media == nil {
		dest.Media = &pb.Media{}
	}
	d := dest.Media

	mediaType := mediaTypeToProto(*src.Type)
	d.MediaId = string(src.MediaID)
	d.MediaProjectId = string(src.MediaProjectID)
	d.Url = strPtrToProto(src.URL)
	d.Type = string(mediaType)
	d.CreatedTime = timestamppb.New(src.CreatedTime)
}

func mediaTypeToProto(t media.Type) pb.MediaType {
	switch t {
	case media.Type_Image:
		return pb.MediaType_MEDIA_TYPE_IMAGE
	case media.Type_Video:
		return pb.MediaType_MEDIA_TYPE_VIDEO
	default:
		return pb.MediaType_MEDIA_TYPE_UNSPECIFIED
	}
}

func metaTypeProtoToEntity(t pb.MediaType) media.Type {
	switch t {
	case pb.MediaType_MEDIA_TYPE_IMAGE:
		return media.Type_Image
	case pb.MediaType_MEDIA_TYPE_VIDEO:
		return media.Type_Video
	default:
		return ""
	}
}
