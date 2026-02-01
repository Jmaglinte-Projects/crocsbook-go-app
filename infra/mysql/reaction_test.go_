package mysql_test

import (
	"context"
	"testing"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/project"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/reaction"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/projectsvc"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/reactionsvc"
	"github.com/stretchr/testify/assert"
)

func TestReactionRepository_Find(t *testing.T) {
	db := mysql.SetupTestDB(t)

	repo := mysql.NewReactionRepository(db)

	ctx := context.Background()

	id := reaction.ReactionID("c2a4fea6-7112-11f0-9198-8abfd21201dc")

	entity, err := repo.Find(ctx, id)
	err = mysql.PrettyPrint(entity)

	assert.NoError(t, err)
	assert.NotNil(t, entity)
	assert.Equal(t, id, entity.ReactionID)

}

func TestReactionRepository_Store(t *testing.T) {
	db := mysql.SetupTestDB(t)

	repo := mysql.NewReactionRepository(db)
	projectSvc := mysql.NewProjectService(db)

	ctx := context.Background()
	now := time.Now()

	// PROJECT

	projectEntities, err := projectSvc.List(ctx, project.ListCond{}, projectsvc.ListOption{
		SortKey: projectsvc.ListOptionSortKey_CreatedAt_DESC,
		Size:    1,
	})
	assert.NoError(t, err)
	assert.NotNil(t, projectEntities)
	mysql.PrettyPrint(projectEntities)

	projectEntity := projectEntities[0]
	// END OF PROJECT

	id, err := reaction.NewReactionID()
	assert.NoError(t, err)

	reactionEntity := &reaction.Reaction{}
	reactionEntity.ReactionID = id
	reactionEntity.ReactionProjectID = reaction.ProjectID(projectEntity.ProjectID)

	reactionType := reaction.Type_Like
	reactionEntity.Type = &reactionType
	reactionEntity.CreatedTime = now

	err = repo.Store(ctx, reactionEntity)
	assert.NoError(t, err)
}

func Test_ReactionRepository_Remove(t *testing.T) {
	db := mysql.SetupTestDB(t)

	repo := mysql.NewReactionRepository(db)

	ctx := context.Background()

	err := repo.Remove(ctx, reaction.ReactionID("c2a4fea6-7112-11f0-9198-8abfd21201dc"))
	assert.NoError(t, err)
}

// SERVICE
func TestReactionService_List(t *testing.T) {
	db := mysql.SetupTestDB(t)

	service := mysql.NewReactionService(db)

	ctx := context.Background()

	// reactionID := reaction.ReactionID("c2a4fea6-7112-11f0-9198-8abfd21201dc")
	offset := int64(0)

	cond := reaction.ListCond{
		// ReactionID: &reactionID,
	}

	opt := reactionsvc.ListOption{
		SortKey: reactionsvc.ListOptionSortKey_CreatedAt_ASC,
		Offset:  &offset,
		Size:    10,
	}

	entities, err := service.List(ctx, cond, opt)
	assert.NoError(t, err)
	assert.NotNil(t, entities)

	err = mysql.PrettyPrint(entities)
	assert.NoError(t, err)
}

func TestReactionService_Count(t *testing.T) {
	db := mysql.SetupTestDB(t)

	service := mysql.NewReactionService(db)

	ctx := context.Background()

	cond := reaction.CountCond{}
	opt := reactionsvc.CountOption{}

	count, err := service.Count(ctx, cond, opt)
	assert.NoError(t, err)

	err = mysql.PrettyPrint(count)
	assert.NoError(t, err)
}
