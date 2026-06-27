package post

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	PostID        PostID
	PostProjectID ProjectID
	Content       *string
	Visibility    *Visibility
	CreatedTime   time.Time
	UpdatedTime   *time.Time

	MediaSets         []MediaSet
	PostReactionCount uint64
	HasReacted        bool
}

type PostID string

func NewPostID() (PostID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return PostID(id.String()), nil
}

type ProjectID string

type Visibility string

const (
	Visibility_Public  Visibility = "Public"
	Visibility_Private Visibility = "Private"
)

type MediaSet struct {
	ContentType string
	Content     []byte
}

type PostReactions struct {
	PostReactionID PostReactionID
	PostID         PostID
	UserID         string
	ReactionType   *ReactionType
	CreatedTime    time.Time
}

type PostReactionID string

func NewPostReactionID() (PostReactionID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return PostReactionID(id.String()), nil
}

type ReactionType string

const (
	ReactionType_Like  ReactionType = "Like"
	ReactionType_Heart ReactionType = "Heart"
)

type ListCond struct {
	UserID         *string
	PostID         *PostID
	PostIDs        []PostID
	PostProjectID  *ProjectID
	PostProjectIDs []ProjectID
	CreatedTime    *time.Time

	SortKey PostSortKey
	Size    int64
	Offset  *int64
}

type CountCond ListCond

type ListPostStatsByProjectIdsCond struct {
	CreatedTime *time.Time

	SortKey PostSortKey
	Size    int64
	Offset  *int64
}

type PostSortKey uint

const (
	PostSortKey_CreatedTime_ASC PostSortKey = iota
	PostSortKey_CreatedTime_DESC
)

type ListPostReactionsCond struct {
	PostReactionID *PostReactionID
	PostID         *PostID
	UserID         *string

	// TODO pagination
	SortKey PostReactionSortKey
	Size    int64
	Offset  *int64
}

type CountPostReactionsCond ListPostReactionsCond

type PostReactionSortKey uint

const (
	PostReactionSortKey_CreatedTime_ASC PostReactionSortKey = iota
	PostReactionSortKey_CreatedTime_DESC
)

type PostComment struct {
	CommentID       PostCommentID
	PostID          string
	UserID          string
	ParentCommentID *PostCommentID
	Content         string
	CreatedTime     time.Time
	UpdatedTime     *time.Time
}

type PostCommentID string

func NewPostCommentID() (PostCommentID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return PostCommentID(id.String()), nil
}

type ListPostCommentCond struct {
	PostCommentID *PostCommentID
	PostID        *PostID
	UserID        *string

	SortKey PostCommentSortKey
	Size    int64
	Offset  *int64
}

type CountPostCommentCond ListPostCommentCond

type PostCommentSortKey uint

const (
	PostCommentSortKey_CreatedTime_ASC PostCommentSortKey = iota
	PostCommentSortKey_CreatedTime_DESC
)
