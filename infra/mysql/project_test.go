package mysql_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/project"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/projectsvc"
	"github.com/stretchr/testify/assert"
)

func TestProjectRepository_Find(t *testing.T) {
	db := mysql.SetupTestDB(t)

	logger := slog.With("Testing", "room_repository")
	repo := mysql.NewProjectRepository(db)

	ctx := context.Background()

	id := project.ProjectID("c2a4fea6-7112-11f0-9198-8abfd21201dc")

	entity, err := repo.Find(ctx, id)
	logger.Info("TestProjectRepository_Find", "entity", entity)
	assert.NoError(t, err)
	assert.NotNil(t, entity)
	assert.Equal(t, id, entity.ProjectID)
}

func TestProjectRepository_Store(t *testing.T) {
	db := mysql.SetupTestDB(t)

	repo := mysql.NewProjectRepository(db)

	ctx := context.Background()
	now := time.Now()

	id := project.ProjectID("c2a4fea6-7112-11f0-9198-8abfd21201dc")
	userID := project.UserID("c2a4fea6-1234-11f0-9198-8abfd21201d2")
	cost := int64(69000000000)

	entity := &project.Project{
		ProjectID:      id,
		ProjectUserID:  userID,
		Name:           "Bulacan",
		Description:    nil,
		Thumbnail:      nil,
		Location:       nil,
		Cost:           &cost,
		StartDate:      &now,
		CreatedTime:    now,
		CompletionDate: &now,
	}

	err := repo.Store(ctx, entity)
	assert.NoError(t, err)
}

// SERVICE
func TestProjectService_List(t *testing.T) {
	db := mysql.SetupTestDB(t)

	service := mysql.NewProjectService(db)

	ctx := context.Background()

	// roomID := reservation.ProjectID("5d641e94-7119-11f0-8845-8abfd21201dc")
	offset := int64(0)

	cond := project.ListCond{
		// ProjectID: &roomID,
	}

	opt := projectsvc.ListOption{
		SortKey: projectsvc.ListOptionSortKey_CreatedAt_ASC,
		Offset:  &offset,
		Size:    10,
	}

	entities, err := service.List(ctx, cond, opt)
	assert.NoError(t, err)
	assert.NotNil(t, entities)

	err = mysql.PrettyPrint(entities)
	assert.NoError(t, err)
}

func TestProjectService_Count(t *testing.T) {
	db := mysql.SetupTestDB(t)

	service := mysql.NewProjectService(db)

	ctx := context.Background()

	cond := project.CountCond{}
	opt := projectsvc.CountOption{}

	count, err := service.Count(ctx, cond, opt)
	assert.NoError(t, err)

	err = mysql.PrettyPrint(count)
	assert.NoError(t, err)
}
