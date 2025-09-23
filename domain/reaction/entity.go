package reaction

import "time"

type Reaction struct {
	ReactionID    string
	PostProjectID string
	Type          *Type
	CreatedTime   time.Time
}

type Type string

var (
	Type_Like Type = "Like"
)
