package media

import (
	"time"

	"github.com/google/uuid"
)

type Media struct {
	MediaID     MediaID
	MediaPostID PostID
	ObjectKey   string
	Type        Type
	CreatedTime time.Time

	MediaSet MediaSet
}

type MediaID string

func NewMediaID() (MediaID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return MediaID(id.String()), nil
}

type PostID string

type Type string

type MediaSet struct {
	ContentType string
	Content     []byte
}

type ListCond struct {
	MediaID  *MediaID
	MediaIDs []MediaID

	MediaPostID *PostID
	// MediaPostIDs []PostID
}

type CountCond ListCond
