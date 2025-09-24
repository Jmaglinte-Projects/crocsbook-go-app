package reaction

import (
	"time"

	"github.com/google/uuid"
)

type Reaction struct {
	ReactionID        ReactionID
	ReactionProjectID ProjectID
	Type              *Type
	CreatedTime       time.Time
}

type ReactionID string

func NewReactionID() (ReactionID, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return ReactionID(id.String()), nil
}

type ProjectID string

type Type string

const (
	Type_Like Type = "Like"
)

type ListCond struct {
	ReactionID  *ReactionID
	ReactionIDs []ReactionID
}

type CountCond ListCond
