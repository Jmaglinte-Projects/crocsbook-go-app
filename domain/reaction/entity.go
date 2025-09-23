package reaction

import "time"

type Reaction struct {
	ReactionID    ReactionID
	PostProjectID ProjectID
	Type          *Type
	CreatedTime   time.Time
}

type ReactionID string

type ProjectID string

type Type string

var (
	Type_Like Type = "Like"
)
