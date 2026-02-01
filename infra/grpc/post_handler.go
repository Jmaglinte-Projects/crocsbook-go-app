package grpc

import (
	"context"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/post"
	pb "github.com/Jmaglinte-Projects/crocsbook-go-app/infra/grpc/lib"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/postsvc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type postServer struct {
	pb.UnimplementedPostServiceServer
	svc postsvc.Service
}

func NewPostHandler(svc postsvc.Service) pb.PostServiceServer {
	return &postServer{
		svc: svc,
	}
}

func (s *postServer) ShowPosts(ctx context.Context, req *pb.ShowPostsIn) (*pb.ShowPostsOut, error) {
	in := &postsvc.ShowPostsIn{}

	posts, err := s.svc.ShowPosts(ctx, in)
	if err != nil {
		return nil, err
	}

	out := pb.ShowPostsOut{
		Items: make([]*pb.ViewPost, len(posts.Items)),
	}
	for _, post := range posts.Items {
		item := &ViewPost{}
		item.UnmarshalOriginal(post)
		out.Items = append(out.Items, &item.ViewPost)
	}

	return &pb.ShowPostsOut{Items: out.Items, Total: out.Total}, nil
}

func (s *postServer) ShowPost(ctx context.Context, req *pb.ShowPostIn) (*pb.ShowPostOut, error) {
	in := &postsvc.ShowPostIn{
		PostID: post.PostID(req.PostId),
	}

	post, err := s.svc.ShowPost(ctx, in)
	if err != nil {
		return nil, err
	}

	out := pb.ShowPostOut{
		Item: &pb.ViewPost{},
	}
	item := &ViewPost{}
	item.UnmarshalOriginal(post.Item)
	out.Item = &item.ViewPost

	return &out, nil
}

func (s *postServer) CreatePost(ctx context.Context, req *pb.CreatePostIn) (*pb.CreatePostOut, error) {
	mediaImages := make([]*postsvc.MediaImage, len(req.MediaImages))
	for i, mediaImage := range req.MediaImages {
		mediaImages[i] = rpcPostMediaImageToSvcMediaImage(mediaImage)
	}

	visibility := postVisibilityToEntity(req.Visibility)
	in := &postsvc.CreatePostIn{
		PostProjectID: post.ProjectID(req.PostProjectId),
		Content:       &req.Content,
		Visibility:    &visibility,
		MediaImages:   &mediaImages,
	}

	_, err := s.svc.CreatePost(ctx, in)
	if err != nil {
		return nil, err
	}

	return &pb.CreatePostOut{}, nil
}

func (s *postServer) UpdatePost(ctx context.Context, req *pb.UpdatePostIn) (*pb.UpdatePostOut, error) {
	visibility := postVisibilityToEntity(req.Visibility)
	in := &postsvc.UpdatePostIn{
		PostID:        post.PostID(req.PostId),
		PostProjectID: post.ProjectID(req.PostProjectId),
		Content:       &req.Content,
		Visibility:    &visibility,
	}
	_, err := s.svc.UpdatePost(ctx, in)
	if err != nil {
		return nil, err
	}

	return &pb.UpdatePostOut{}, nil
}

func (s *postServer) RemovePost(ctx context.Context, req *pb.RemovePostIn) (*pb.RemovePostOut, error) {
	in := &postsvc.RemovePostIn{
		PostID: post.PostID(req.PostId),
	}
	_, err := s.svc.RemovePost(ctx, in)
	if err != nil {
		return nil, err
	}

	return &pb.RemovePostOut{}, nil
}

type ViewPost struct {
	pb.ViewPost
}

func (dest *ViewPost) UnmarshalOriginal(src *postsvc.ViewPost) {
	if dest.Post == nil {
		dest.Post = &pb.Post{}
	}
	d := dest.Post

	d.PostId = string(src.PostID)
	d.PostProjectId = string(src.PostProjectID)
	d.Content = strPtr(src.Content)
	d.Visibility = postVisibilityToProto(*src.Visibility)
	d.CreatedTime = timestamppb.New(src.CreatedTime)
	d.UpdatedTime = timePtrToProto(src.UpdatedTime)
}

func postVisibilityToProto(v post.Visibility) pb.PostVisibility {
	switch v {
	case post.Visibility_Public:
		return pb.PostVisibility_PostVisibility_Public
	case post.Visibility_Private:
		return pb.PostVisibility_PostVisibility_Private
	default:
		return pb.PostVisibility_PostVisibility_UNSPECIFIED
	}
}

func postVisibilityToEntity(v pb.PostVisibility) post.Visibility {
	switch v {
	case pb.PostVisibility_PostVisibility_Public:
		return post.Visibility_Public
	case pb.PostVisibility_PostVisibility_Private:
		return post.Visibility_Private
	default:
		return ""
	}
}

func rpcPostMediaImageToSvcMediaImage(src *pb.MediaImage) *postsvc.MediaImage {
	return &postsvc.MediaImage{
		Filename:    src.Filename,
		ContentType: src.ContentType,
		Content:     src.Content,
	}
}
