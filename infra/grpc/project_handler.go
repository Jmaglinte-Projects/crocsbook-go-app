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
		Items: make([]*pb.ViewProject, len(projects.Items)),
	}

	for _, project := range projects.Items {
		item := &VieWProject{}
		item.UnmarshalOriginal(project)
		out.Items = append(out.Items, &item.ViewProject)

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
		Item: &pb.ViewProject{},
	}

	item := &VieWProject{}
	item.UnmarshalOriginal(project.Item)
	out.Item = &item.ViewProject

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

type VieWProject struct {
	pb.ViewProject
}

func (dest *VieWProject) UnmarshalOriginal(src *projectsvc.ViewProject) {
	if dest.Project == nil {
		dest.Project = &pb.Project{}
	}
	d := dest.Project

	d.ProjectId = string(src.ProjectID)
	d.ProjectUserId = string(src.ProjectUserID)
	d.Name = src.Name

	if src.Description != nil {
		d.Description = *src.Description
	}

	if src.Thumbnail != nil {
		d.Thumbnail = *src.Thumbnail
	}

	if src.Location != nil {
		d.Location = *src.Location
	}

	if src.Cost != nil {
		d.Cost = *src.Cost
	}

	if src.StartDate != nil {
		d.StartDate = timestamppb.New(*src.StartDate)
	}

	if src.CompletionDate != nil {
		d.CompletionDate = timestamppb.New(*src.CompletionDate)
	}

	d.CreatedTime = timestamppb.New(src.CreatedTime)

	if src.UpdatedTime != nil {
		d.UpdatedTime = timestamppb.New(*src.UpdatedTime)
	}

	dest.ThumbnailUrl = src.ThumbnailURL

}
