package media

import "time"

type Media struct {
	MediaID        MediaID
	MediaProjectID ProjectID
	URL            *string
	Type           *Type
	CreatedTime    time.Time
}

type MediaID string

type ProjectID string

type Type string

const (
	Type_Image Type = "Image"
	Type_Video Type = "Video"
)
