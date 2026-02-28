package grpc

import (
	"context"
	"fmt"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/project"
	pb "github.com/Jmaglinte-Projects/crocsbook-go-app/infra/grpc/lib"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/projectsvc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type projectServer struct {
	pb.UnimplementedProjectServiceServer
	svc projectsvc.Service
}

func NewProjectHandler(svc projectsvc.Service) pb.ProjectServiceServer {
	return &projectServer{
		svc: svc,
	}
}

func (s *projectServer) ShowProjects(ctx context.Context, req *pb.ShowProjectsIn) (*pb.ShowProjectsOut, error) {
	in := projectsvc.ShowProjectsIn{}

	projects, err := s.svc.ShowProjects(ctx, &in)
	if err != nil {
		return nil, err
	}

	out := pb.ShowProjectsOut{
		Items: make([]*pb.Project, 0, len(projects.Items)),
	}

	for _, project := range projects.Items {
		item := &Project{}
		item.UnmarshalOriginal(project)
		out.Items = append(out.Items, (*pb.Project)(item))
	}

	return &out, nil
}

func (s *projectServer) ShowProject(ctx context.Context, req *pb.ShowProjectIn) (*pb.ShowProjectOut, error) {
	in := projectsvc.ShowProjectIn{
		ProjectID: project.ProjectID(req.ProjectId),
	}

	project, err := s.svc.ShowProject(ctx, &in)
	if err != nil {
		return nil, err
	}

	out := pb.ShowProjectOut{
		Item: &pb.Project{},
	}

	item := &Project{}
	item.UnmarshalOriginal(project.Item)
	out.Item = (*pb.Project)(item) //type conversion

	return &out, nil
}

func (s *projectServer) CreateProject(ctx context.Context, req *pb.CreateProjectIn) (*pb.CreateProjectOut, error) {
	startDate := req.StartDate.AsTime()
	completionDate := req.CompletionDate.AsTime()

	fmt.Println("ProjectUserId: ", req.ProjectUserId)

	in := projectsvc.CreateProjectIn{
		ProjectUserID:    project.UserID(req.ProjectUserId),
		Name:             req.Name,
		Description:      &req.Description,
		ThumbnailContent: req.ThumbnailContent,
		Location:         &req.Location,
		Cost:             &req.Cost,
		StartDate:        &startDate,
		CompletionDate:   &completionDate,
	}

	fmt.Println("--------------------------------")
	fmt.Println("in.ThumbnailContent: ", in.ThumbnailContent)
	fmt.Println("--------------------------------")

	_, err := s.svc.CreateProject(ctx, &in)
	if err != nil {
		return nil, err
	}

	return &pb.CreateProjectOut{}, nil
}

func (s *projectServer) UpdateProject(ctx context.Context, req *pb.UpdateProjectIn) (*pb.UpdateProjectOut, error) {
	startDate := req.StartDate.AsTime()
	completionDate := req.CompletionDate.AsTime()
	in := projectsvc.UpdateProjectIn{
		ProjectID:        project.ProjectID(req.ProjectId),
		Name:             req.Name,
		Description:      &req.Description,
		ThumbnailContent: req.ThumbnailContent,
		Location:         &req.Location,
		Cost:             &req.Cost,
		StartDate:        &startDate,
		CompletionDate:   &completionDate,
	}

	_, err := s.svc.UpdateProject(ctx, &in)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateProjectOut{}, nil
}

func (s *projectServer) RemoveProject(ctx context.Context, req *pb.RemoveProjectIn) (*pb.RemoveProjectOut, error) {
	in := projectsvc.RemoveProjectIn{
		ProjectID: project.ProjectID(req.ProjectId),
	}

	_, err := s.svc.RemoveProject(ctx, &in)
	if err != nil {
		return nil, err
	}

	return &pb.RemoveProjectOut{}, nil
}

type Project pb.Project

func (dest *Project) UnmarshalOriginal(src *projectsvc.ViewProject) {
	dest.ProjectId = string(src.ProjectID)
	dest.ProjectUserId = string(src.ProjectUserID)
	dest.Name = src.Name

	if src.Description != nil {
		dest.Description = *src.Description
	}

	if src.ThumbnailKey != nil {
		dest.ThumbnailKey = *src.ThumbnailKey
	}

	if src.Location != nil {
		dest.Location = *src.Location
	}

	if src.Cost != nil {
		dest.Cost = *src.Cost
	}

	if src.StartDate != nil {
		dest.StartDate = timestamppb.New(*src.StartDate)
	}

	if src.CompletionDate != nil {
		dest.CompletionDate = timestamppb.New(*src.CompletionDate)
	}

	dest.CreatedTime = timestamppb.New(src.CreatedTime)

	if src.UpdatedTime != nil {
		dest.UpdatedTime = timestamppb.New(*src.UpdatedTime)
	}

	dest.ThumbnailUrl = src.ThumbnailURL

}
