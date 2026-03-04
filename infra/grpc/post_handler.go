package grpc

import (
	"context"
	"encoding/json"
	"fmt"

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
	postFilter := &PostFilter{}
	postFilter.Marshal(req.Filter)
	filterSvc := postsvc.Filter(*postFilter)

	in := &postsvc.ShowPostsIn{
		Filter: &filterSvc,
	}

	b, _ := json.MarshalIndent(req.Filter, "", "  ")
	fmt.Println("in.Filter:", string(b))

	posts, err := s.svc.ShowPosts(ctx, in)
	if err != nil {
		return nil, err
	}

	out := pb.ShowPostsOut{
		Items: make([]*pb.Post, 0, len(posts.Items)),
		Total: posts.Total,
	}

	for _, p := range posts.Items {
		item := &Post{}
		item.UnmarshalOriginal(p)
		out.Items = append(out.Items, &item.Post)
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

	item := &Post{}
	item.UnmarshalOriginal(post.Item)
	out := pb.ShowPostOut{
		Item: &item.Post,
	}
	return &out, nil
}

func (s *postServer) CreatePost(ctx context.Context, req *pb.CreatePostIn) (*pb.CreatePostOut, error) {
	visibility := postVisibilityToEntity(req.Visibility)
	in := &postsvc.CreatePostIn{
		PostProjectID: post.ProjectID(req.PostProjectId),
		Content:       &req.Content,
		Visibility:    &visibility,
		MediaList:     req.Images,
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

func (s *postServer) ShowPostByProjectId(ctx context.Context, req *pb.ShowPostByProjectIdIn) (*pb.ShowPostByProjectIdOut, error) {
	filter := &PostFilter{}
	filter.Marshal(req.Filter)
	filterSvc := postsvc.Filter(*filter)

	// b, _ := json.MarshalIndent(filter, "", "  ")
	// fmt.Println("postFilter:", string(b))

	in := &postsvc.ShowPostByProjectIdIn{
		ProjectID: post.ProjectID(req.ProjectId),
		Filter:    &filterSvc,
	}

	posts, err := s.svc.ShowPostByProjectId(ctx, in)
	if err != nil {
		return nil, err
	}

	out := pb.ShowPostsOut{
		Items: make([]*pb.Post, 0, len(posts.Items)),
		Total: posts.Total,
	}

	for _, p := range posts.Items {
		item := &Post{}
		item.UnmarshalOriginal(p)
		out.Items = append(out.Items, &item.Post)
	}

	return &pb.ShowPostByProjectIdOut{Items: out.Items, Total: out.Total}, nil
}

type Post struct {
	pb.Post
}

func (dest *Post) UnmarshalOriginal(src *postsvc.ViewPost) {
	dest.PostId = string(src.PostID)
	dest.PostProjectId = string(src.PostProjectID)
	if src.Content != nil {
		dest.Content = *src.Content
	}
	if src.Visibility != nil {
		dest.Visibility = postVisibilityToProto(*src.Visibility)
	}
	dest.CreatedTime = timestamppb.New(src.CreatedTime)
	if src.UpdatedTime != nil {
		dest.UpdatedTime = timestamppb.New(*src.UpdatedTime)
	}

	dest.PostCount = src.PostCount
	dest.LastPostTime = timestamppb.New(src.LastPostTime)
	dest.PostReactionCount = src.PostReactionCount
	dest.HasReacted = src.HasReacted

	if src.Project != nil {
		dest.Project = &pb.Post_Project{}
		dest.Project.ProjectId = string(src.Project.ProjectID)
		dest.Project.ProjectUserId = string(src.Project.ProjectUserID)
		dest.Project.Name = src.Project.Name
		if src.Project.Description != nil {
			dest.Project.Description = *src.Project.Description
		}
		if src.Project.ThumbnailKey != nil {
			dest.Project.ThumbnailKey = *src.Project.ThumbnailKey
		}
		if src.Project.Location != nil {
			dest.Project.Location = *src.Project.Location
		}
		if src.Project.Cost != nil {
			dest.Project.Cost = *src.Project.Cost
		}
		if src.Project.StartDate != nil {
			dest.Project.StartDate = timestamppb.New(*src.Project.StartDate)
		}
		if src.Project.CompletionDate != nil {
			dest.Project.CompletionDate = timestamppb.New(*src.Project.CompletionDate)
		}
		dest.Project.CreatedTime = timestamppb.New(src.Project.CreatedTime)
		if src.Project.UpdatedTime != nil {
			dest.Project.UpdatedTime = timestamppb.New(*src.Project.UpdatedTime)
		}
		dest.Project.ThumbnailUrl = src.Project.ThumbnailURL
	}
	if src.MediaList != nil {
		dest.MediaList = make([]*pb.Post_Media, 0, len(src.MediaList))
		for _, m := range src.MediaList {
			dest.MediaList = append(dest.MediaList, &pb.Post_Media{
				MediaId:      string(m.MediaID),
				MediaPostId:  string(m.MediaPostID),
				ObjectKey:    m.ObjectKey,
				Type:         string(m.Type),
				CreatedTime:  timestamppb.New(m.CreatedTime),
				ThumbnailUrl: m.PresignedURL,
			})
		}
	}
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

type PostFilter postsvc.Filter

func (dest *PostFilter) Marshal(src *pb.PostFilter) {
	if src.CreatedTime != nil {
		t := src.CreatedTime.AsTime()
		dest.CreatedTime = &t
	}

	dest.SortKey = post.PostSortKey(src.SortKey)
	dest.Size = src.Size
	dest.Offset = src.Offset
}

// func rpcPostMediaImageToSvcMediaImage(src *pb.MediaImage) *postsvc.MediaImage {
// 	return &postsvc.MediaImage{
// 		Filename:    src.Filename,
// 		ContentType: src.ContentType,
// 		Content:     src.Content,
// 	}
// }
