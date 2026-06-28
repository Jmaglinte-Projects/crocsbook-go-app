package postsvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/media"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/post"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/project"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/user"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/mediasvc"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/projectsvc"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/usersvc"
	"github.com/gabriel-vasile/mimetype"
)

type PostRepository interface {
	Find(ctx context.Context, id post.PostID) (*ViewPost, error)
	Store(ctx context.Context, entity *post.Post) error
	Remove(ctx context.Context, ids ...post.PostID) error
}

type PostReactionRepository interface {
	Find(ctx context.Context, id post.PostReactionID) (*post.PostReactions, error)
	Store(ctx context.Context, entity *post.PostReactions) error
	Remove(ctx context.Context, ids ...post.PostReactionID) error
}

type PostCommentRepository interface {
	Find(ctx context.Context, id post.PostCommentID) (*post.PostComment, error)
	Store(ctx context.Context, entity *post.PostComment) error
	Remove(ctx context.Context, ids ...post.PostCommentID) error
}

type PostService interface {
	List(ctx context.Context, cond post.ListCond) ([]*ViewPost, error)
	Count(ctx context.Context, cond post.CountCond) (*uint64, error)
	ListPostStatsByProjectIds(ctx context.Context, cond post.ListPostStatsByProjectIdsCond, projectId ...post.ProjectID) ([]*ListPostStatsByProjectIds, error)
}

type PostReactionService interface {
	List(ctx context.Context, cond post.ListPostReactionsCond) ([]*post.PostReactions, error)
	Count(ctx context.Context, cond post.CountPostReactionsCond) (*uint64, error)
}

type PostCommentService interface {
	List(ctx context.Context, cond post.ListPostCommentCond) ([]*post.PostComment, error)
	Count(ctx context.Context, cond post.CountPostCommentCond) (*uint64, error)
}

type Service interface {
	ShowPosts(ctx context.Context, in *ShowPostsIn) (*ShowPostsOut, error)
	ShowPost(ctx context.Context, in *ShowPostIn) (*ShowPostOut, error)
	CreatePost(ctx context.Context, in *CreatePostIn) (*CreatePostOut, error)
	UpdatePost(ctx context.Context, in *UpdatePostIn) (*UpdatePostOut, error)
	RemovePost(ctx context.Context, in *RemovePostIn) (*RemovePostOut, error)

	ReactToPost(ctx context.Context, in *ReactToPostIn) (*ReactToPostOut, error)
	CommentOnPost(ctx context.Context, in *CommentOnPostIn) (*CommentOnPostOut, error)

	ShowPostByProjectId(ctx context.Context, in *ShowPostByProjectIdIn) (*ShowPostByProjectIdOut, error)
}

// Todo: pagination
type ShowPostsIn struct {
	Filter *Filter
}
type ShowPostsOut struct {
	Items []*ViewPost
	Total uint64
}

type ShowPostIn struct {
	PostID post.PostID
}
type ShowPostOut struct {
	Item *ViewPost
}

type CreatePostIn struct {
	PostProjectID post.ProjectID
	Content       *string
	Visibility    *post.Visibility

	MediaList [][]byte
}
type CreatePostOut struct{}

type UpdatePostIn struct {
	PostID        post.PostID
	PostProjectID post.ProjectID

	Content    *string
	Visibility *post.Visibility
}
type UpdatePostOut struct{}

type RemovePostIn struct {
	PostID post.PostID
}
type RemovePostOut struct{}

type ShowPostByProjectIdIn struct {
	ProjectID post.ProjectID

	Filter *Filter
}
type ShowPostByProjectIdOut struct {
	Items []*ViewPost
	Total uint64
}

type ReactToPostIn struct {
	PostID       post.PostID
	UserID       string
	ReactionType post.ReactionType
}

type ReactToPostOut struct{}

type CommentOnPostIn struct {
	PostID          post.PostID
	UserID          string
	Content         string
	ParentCommentID *post.PostCommentID
}

type CommentOnPostOut struct {
	Item *ViewPostComment
}

type service struct {
	postRepo         PostRepository
	postSvc          PostService
	postReactionRepo PostReactionRepository
	postReactionSvc  PostReactionService
	postCommentRepo  PostCommentRepository
	postCommentSvc   PostCommentService
	mediaRepo        mediasvc.MediaRepository
	mediaSvc         mediasvc.MediaService
	projectSvc       projectsvc.ProjectService
	projectR2Repo    projectsvc.ProjectR2Repository
	userSvc          usersvc.UserService
}

func NewService(postRepo PostRepository, postSvc PostService, postReactionRepo PostReactionRepository, postReactionSvc PostReactionService, postCommentRepo PostCommentRepository, postCommentSvc PostCommentService, mediaRepo mediasvc.MediaRepository, mediaSvc mediasvc.MediaService, projectSvc projectsvc.ProjectService, projectR2Repo projectsvc.ProjectR2Repository, userSvc usersvc.UserService) Service {
	return &service{
		postRepo:         postRepo,
		postSvc:          postSvc,
		postReactionRepo: postReactionRepo,
		postReactionSvc:  postReactionSvc,
		postCommentRepo:  postCommentRepo,
		postCommentSvc:   postCommentSvc,
		mediaRepo:        mediaRepo,
		mediaSvc:         mediaSvc,
		projectSvc:       projectSvc,
		projectR2Repo:    projectR2Repo,
		userSvc:          userSvc,
	}
}

func (s *service) ShowPosts(ctx context.Context, in *ShowPostsIn) (*ShowPostsOut, error) {
	cond := &post.ListCond{}
	in.Filter.Unmarshal(cond)

	entities, err := s.postSvc.List(ctx, *cond)
	if err != nil {
		return nil, err
	}

	if err := s.setProject(ctx, entities...); err != nil {
		return nil, err
	}

	if err := s.setMedia(ctx, entities...); err != nil {
		return nil, err
	}

	if err := s.setPostStatsByProjectIds(ctx, entities...); err != nil {
		return nil, err
	}

	if err := s.setUser(ctx, entities...); err != nil {
		return nil, err
	}

	if err := s.setPostComments(ctx, entities...); err != nil {
		return nil, err
	}

	// b, _ := json.MarshalIndent(entities, "", "  ")
	// fmt.Println("entities:", string(b))

	countCond := &post.CountCond{}
	count, err := s.postSvc.Count(ctx, *countCond)
	if err != nil {
		return nil, err
	}

	return &ShowPostsOut{
		Items: entities,
		Total: *count,
	}, nil
}

func (s *service) ShowPost(ctx context.Context, in *ShowPostIn) (*ShowPostOut, error) {
	entity, err := s.postRepo.Find(ctx, post.PostID(in.PostID))
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, errors.New("Entity not found")
	}

	if err := s.setMedia(ctx, entity); err != nil {
		return nil, err
	}

	if err := s.setProject(ctx, entity); err != nil {
		return nil, err
	}

	if err := s.setUser(ctx, entity); err != nil {
		return nil, err
	}

	if err := s.setPostComments(ctx, entity); err != nil {
		return nil, err
	}

	// b, _ := json.MarshalIndent(entity.Post, "", "  ")
	// fmt.Println("entity.Post:", string(b))

	return &ShowPostOut{
		Item: entity,
	}, nil
}

func (s *service) CreatePost(ctx context.Context, in *CreatePostIn) (*CreatePostOut, error) {
	now := time.Now()

	id, err := post.NewPostID()
	if err != nil {
		return nil, err
	}

	entity := &post.Post{}
	entity.PostID = id
	entity.PostProjectID = in.PostProjectID
	entity.Content = in.Content
	entity.Visibility = in.Visibility
	entity.CreatedTime = now

	err = s.postRepo.Store(ctx, entity)
	if err != nil {
		return nil, err
	}

	for _, mediaItem := range in.MediaList {
		createMediaIn := &mediasvc.CreateMediaIn{
			MediaPostID: media.PostID(entity.PostID),
			Data:        mediaItem,
		}
		_, err = s.createMedia(ctx, createMediaIn)
		if err != nil {
			return nil, err
		}
	}

	return &CreatePostOut{}, nil
}

func (s *service) UpdatePost(ctx context.Context, in *UpdatePostIn) (*UpdatePostOut, error) {
	now := time.Now()

	entity, err := s.postRepo.Find(ctx, in.PostID)
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, errors.New("Entity not found")
	}

	entity.Post.Content = in.Content
	entity.Visibility = in.Visibility
	entity.UpdatedTime = &now

	err = s.postRepo.Store(ctx, &entity.Post)
	if err != nil {
		return nil, err
	}

	// TODO: update media feature

	return &UpdatePostOut{}, nil
}

func (s *service) RemovePost(ctx context.Context, in *RemovePostIn) (*RemovePostOut, error) {
	entity, err := s.postRepo.Find(ctx, in.PostID)
	if err != nil {
		return nil, err
	}

	if entity == nil {
		return nil, errors.New("Entity not found")
	}

	err = s.postRepo.Remove(ctx, entity.PostID)
	if err != nil {
		return nil, err
	}

	return &RemovePostOut{}, nil
}

func (s *service) ReactToPost(ctx context.Context, in *ReactToPostIn) (*ReactToPostOut, error) {
	now := time.Now()

	hasReacted, err := s.userHasReactedToPost(ctx, in.UserID, in.PostID)
	fmt.Println("hasReacted:", hasReacted)
	if err != nil {
		return nil, err
	}
	if hasReacted {
		err = s.unReactToPost(ctx, in.UserID, in.PostID)
		fmt.Println("unReactToPost step 0")
		if err != nil {
			fmt.Println("Error unreacting to post:", err)
			return nil, err
		}
		return &ReactToPostOut{}, nil
	}

	id, err := post.NewPostReactionID()
	if err != nil {
		return nil, err
	}

	entity := &post.PostReactions{}
	entity.PostReactionID = id
	entity.PostID = in.PostID
	entity.UserID = in.UserID
	entity.ReactionType = &in.ReactionType
	entity.CreatedTime = now

	b, _ := json.MarshalIndent(entity, "", "  ")
	fmt.Println("ReactToPost entity:", string(b))

	err = s.postReactionRepo.Store(ctx, entity)
	if err != nil {
		fmt.Println("Error storing post reaction:", err)
		return nil, err
	}

	return &ReactToPostOut{}, nil
}

func (s *service) CommentOnPost(ctx context.Context, in *CommentOnPostIn) (*CommentOnPostOut, error) {
	if in.Content == "" {
		return nil, errors.New("comment content is required")
	}

	postEntity, err := s.postRepo.Find(ctx, in.PostID)
	if err != nil {
		return nil, err
	}
	if postEntity == nil {
		return nil, errors.New("post not found")
	}

	if in.ParentCommentID != nil {
		parent, err := s.postCommentRepo.Find(ctx, *in.ParentCommentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, errors.New("parent comment not found")
		}
		if parent.PostID != string(in.PostID) {
			return nil, errors.New("parent comment does not belong to this post")
		}
	}

	now := time.Now()

	id, err := post.NewPostCommentID()
	if err != nil {
		return nil, err
	}

	entity := &post.PostComment{}
	entity.CommentID = id
	entity.PostID = string(in.PostID)
	entity.UserID = in.UserID
	entity.ParentCommentID = in.ParentCommentID
	entity.Content = in.Content
	entity.CreatedTime = now

	err = s.postCommentRepo.Store(ctx, entity)
	if err != nil {
		return nil, err
	}

	view := &ViewPostComment{PostComment: *entity}
	if err := s.setCommentUsers(ctx, view); err != nil {
		return nil, err
	}

	return &CommentOnPostOut{Item: view}, nil
}

func (s *service) ShowPostByProjectId(ctx context.Context, in *ShowPostByProjectIdIn) (*ShowPostByProjectIdOut, error) {
	cond := &post.ListCond{
		PostProjectID: &in.ProjectID,
	}
	// for filters
	in.Filter.Unmarshal(cond)

	b, _ := json.MarshalIndent(cond, "", "  ")
	fmt.Println("ShowPostByProjectId cond:", string(b))

	entities, err := s.postSvc.List(ctx, *cond)
	if err != nil {
		return nil, err
	}

	if err := s.setMedia(ctx, entities...); err != nil {
		return nil, err
	}

	if err := s.setProject(ctx, entities...); err != nil {
		return nil, err
	}

	if err := s.setPostStatsByProjectIds(ctx, entities...); err != nil {
		return nil, err
	}

	if err := s.setUser(ctx, entities...); err != nil {
		return nil, err
	}

	if err := s.setPostComments(ctx, entities...); err != nil {
		return nil, err
	}

	return &ShowPostByProjectIdOut{
		Items: entities,
		Total: uint64(len(entities)),
	}, nil
}

func (s *service) unReactToPost(ctx context.Context, userID string, postID post.PostID) error {
	cond := &post.ListPostReactionsCond{
		UserID: &userID,
		PostID: &postID,
	}
	entities, err := s.postReactionSvc.List(ctx, *cond)
	if err != nil {
		return err
	}
	if len(entities) == 0 {
		return errors.New("user has not reacted to this post")
	}

	fmt.Println("unReactToPost step 1")
	b, _ := json.MarshalIndent(entities[0], "", "  ")
	fmt.Println("unReactToPost entities:", string(b))

	err = s.postReactionRepo.Remove(ctx, entities[0].PostReactionID)
	fmt.Println("unReactToPost step 2")
	if err != nil {
		return err
	}
	return nil
}

func (s *service) userHasReactedToPost(ctx context.Context, userID string, postID post.PostID) (bool, error) {
	cond := &post.ListPostReactionsCond{
		UserID: &userID,
		PostID: &postID,
	}
	entities, err := s.postReactionSvc.List(ctx, *cond)
	if err != nil {
		return false, err
	}
	return len(entities) > 0, nil
}

func (s *service) createMedia(ctx context.Context, in *mediasvc.CreateMediaIn) (*mediasvc.CreateMediaOut, error) {
	now := time.Now()

	id, err := media.NewMediaID()
	if err != nil {
		return nil, err
	}

	entity := &media.Media{}
	entity.MediaID = id
	entity.MediaPostID = in.MediaPostID

	// entity.Type = in.Type
	mt := mimetype.Detect(in.Data)
	entity.MediaSet.Content = in.Data
	entity.MediaSet.ContentType = mt.String()
	entity.CreatedTime = now

	if err = s.mediaRepo.Store(ctx, entity); err != nil {
		return nil, err
	}

	return &mediasvc.CreateMediaOut{}, nil
}

func (s *service) setMedia(ctx context.Context, entities ...*ViewPost) error {
	for _, entity := range entities {
		postID := media.PostID(entity.PostID)
		mediaCond := &media.ListCond{
			MediaPostID: &postID,
		}

		mediaOpt := &mediasvc.ListOption{}
		mediaList, err := s.mediaSvc.List(ctx, *mediaCond, *mediaOpt)
		if err != nil {
			fmt.Println("Error listing media")
			return err
		}

		entity.MediaList = append(entity.MediaList, mediaList...)
	}
	return nil
}

func (s *service) setProject(ctx context.Context, entities ...*ViewPost) error {
	unique := make(map[project.ProjectID]struct{}, len(entities))
	for _, entity := range entities {
		id := project.ProjectID(entity.PostProjectID)
		unique[id] = struct{}{}
	}
	projectIDs := make([]project.ProjectID, 0, len(unique))
	for id := range unique {
		projectIDs = append(projectIDs, id)
	}

	projectCond := &project.ListCond{
		ProjectIDs: projectIDs,
	}
	projectOpt := &projectsvc.ListOption{}
	projects, err := s.projectSvc.List(ctx, *projectCond, *projectOpt)
	if err != nil {
		fmt.Println("Error listing projects")
		return err
	}

	projectByID := make(map[project.ProjectID]*projectsvc.ViewProject)
	for _, p := range projects {
		projectByID[p.ProjectID] = p
	}
	for _, entity := range entities {
		if p, ok := projectByID[project.ProjectID(entity.PostProjectID)]; ok {
			url := ""
			if p.ThumbnailKey != nil {
				url = fmt.Sprintf("%s/%s", os.Getenv("R2_PUBLIC_BASE_URL"), *p.ThumbnailKey)
			}
			entity.Project = &ViewProject{
				Project:      p.Project,
				ThumbnailURL: url,
			}
		}
	}
	return nil
}

func (s *service) setUser(ctx context.Context, entities ...*ViewPost) error {
	unique := make(map[user.UserID]struct{}, len(entities))
	for _, entity := range entities {
		if entity.Project == nil {
			continue
		}
		id := user.UserID(entity.Project.ProjectUserID)
		unique[id] = struct{}{}
	}
	if len(unique) == 0 {
		return nil
	}
	userIDs := make([]user.UserID, 0, len(unique))
	for id := range unique {
		userIDs = append(userIDs, id)
	}

	userCond := &user.ListCond{
		UserIDs: userIDs,
	}
	userOpt := &usersvc.ListOption{}
	users, err := s.userSvc.List(ctx, *userCond, *userOpt)
	if err != nil {
		fmt.Println("Error listing users")
		return err
	}

	userByID := make(map[user.UserID]*usersvc.ViewUser)
	for _, u := range users {
		userByID[u.UserID] = u
	}
	for _, entity := range entities {
		if entity.Project == nil {
			continue
		}
		if u, ok := userByID[user.UserID(entity.Project.ProjectUserID)]; ok {
			s.setProfileUrl(u)
			entity.User = u
		}
	}
	return nil
}

func (s *service) setProfileUrl(entity *usersvc.ViewUser) {
	if entity.User.ProfileKey != nil {
		url := fmt.Sprintf("%s/%s", os.Getenv("R2_PUBLIC_BASE_URL"), *entity.User.ProfileKey)
		entity.User.ProfileURL = &url
	}
}

func (s *service) setPostStatsByProjectIds(ctx context.Context, entities ...*ViewPost) error {
	unique := make(map[post.ProjectID]struct{}, len(entities))
	for _, entity := range entities {
		id := post.ProjectID(entity.PostProjectID)
		unique[id] = struct{}{}
	}
	projectIDs := make([]post.ProjectID, 0, len(unique))
	for id := range unique {
		projectIDs = append(projectIDs, id)
	}

	cond := &post.ListPostStatsByProjectIdsCond{}
	out, err := s.postSvc.ListPostStatsByProjectIds(ctx, *cond, projectIDs...)
	if err != nil {
		fmt.Println("Error counting total posts by project ID")
		return err
	}

	projectByID := make(map[project.ProjectID]*ListPostStatsByProjectIds)
	for _, p := range out {
		projectByID[project.ProjectID(p.ProjectID)] = p
	}

	for _, entity := range entities {
		if p, ok := projectByID[project.ProjectID(entity.PostProjectID)]; ok {
			entity.PostCount = p.Count
			entity.LastPostTime = p.LastPostTime
		}
	}

	return nil
}

func (s *service) setPostComments(ctx context.Context, entities ...*ViewPost) error {
	for _, entity := range entities {
		postID := entity.PostID
		cond := post.ListPostCommentCond{
			PostID:  &postID,
			SortKey: post.PostCommentSortKey_CreatedTime_ASC,
		}

		comments, err := s.postCommentSvc.List(ctx, cond)
		if err != nil {
			return err
		}

		viewComments := make([]*ViewPostComment, 0, len(comments))
		for _, c := range comments {
			viewComments = append(viewComments, &ViewPostComment{PostComment: *c})
		}

		if err := s.setCommentUsers(ctx, viewComments...); err != nil {
			return err
		}

		entity.CommentList = viewComments
		entity.CommentCount = uint64(len(comments))
	}
	return nil
}

func (s *service) setCommentUsers(ctx context.Context, comments ...*ViewPostComment) error {
	unique := make(map[user.UserID]struct{}, len(comments))
	for _, c := range comments {
		id := user.UserID(c.UserID)
		unique[id] = struct{}{}
	}
	if len(unique) == 0 {
		return nil
	}

	userIDs := make([]user.UserID, 0, len(unique))
	for id := range unique {
		userIDs = append(userIDs, id)
	}

	userCond := &user.ListCond{UserIDs: userIDs}
	userOpt := &usersvc.ListOption{}
	users, err := s.userSvc.List(ctx, *userCond, *userOpt)
	if err != nil {
		return err
	}

	userByID := make(map[user.UserID]*usersvc.ViewUser)
	for _, u := range users {
		userByID[u.UserID] = u
	}

	for _, c := range comments {
		if u, ok := userByID[user.UserID(c.UserID)]; ok {
			s.setProfileUrl(u)
			c.User = u
		}
	}
	return nil
}

type Filter struct {
	CreatedTime *time.Time

	SortKey post.PostSortKey
	Size    int64
	Offset  *int64
}

func (src *Filter) Unmarshal(dest *post.ListCond) {
	if src.CreatedTime != nil {
		dest.CreatedTime = src.CreatedTime
	}

	dest.SortKey = src.SortKey
	dest.Size = src.Size

	if src.Offset != nil {
		dest.Offset = src.Offset
	}
}

type ViewPost struct {
	post.Post

	// linked other domain here whenever you need them
	MediaList []*mediasvc.ViewMedia // it should be under postsvc not on mediasvc refactor this later
	Project   *ViewProject
	User      *usersvc.ViewUser

	PostCount    uint64
	LastPostTime time.Time

	CommentList  []*ViewPostComment
	CommentCount uint64
}

type ViewPostComment struct {
	post.PostComment

	User *usersvc.ViewUser
}

type MediaImage struct {
	Filename    string
	ContentType string
	Content     []byte
}

type ViewProject struct {
	project.Project

	ThumbnailURL string
}

type ListPostStatsByProjectIds struct {
	ProjectID    post.ProjectID
	Count        uint64
	LastPostTime time.Time
}
