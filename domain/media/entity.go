package media

import "time"

type Media struct {
	MediaID        string `sql:"primary_key"`
	MediaProjectID string `sql:"primary_key"`
	URL            *string
	Type           *Type
	CreatedTime    time.Time
}

type Type string

const (
	Type_Image Type = "Image"
	Type_Video Type = "Video"
)
