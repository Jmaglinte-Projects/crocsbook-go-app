package media

import (
	"time"

	"github.com/google/uuid"
)

type Media struct {
	MediaID        MediaID
	MediaProjectID ProjectID
	URL            *string
	Type           *Type
	CreatedTime    time.Time
}

type MediaID string

func NewMediaID() (MediaID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return MediaID(id.String()), nil
}

type ProjectID string

type Type string

const (
	Type_Image Type = "Image"
	Type_Video Type = "Video"
)

type ListCond struct {
	MediaID  *MediaID
	MediaIDs []MediaID

	MediaProjectID  *ProjectID
	MediaProjectIDs []ProjectID
}

type CountCond ListCond
