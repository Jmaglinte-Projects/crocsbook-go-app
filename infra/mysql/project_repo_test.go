package mysql_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/project"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql"
	"github.com/stretchr/testify/assert"
)

func TestReservationRepository_Find(t *testing.T) {
	db := mysql.SetupTestDB(t)

	logger := slog.With("Testing", "room_repository")
	repo := mysql.NewProjectRepository(db)

	ctx := context.Background()

	id := project.ProjectID("c2a4fea6-7112-11f0-9198-8abfd21201dc")

	entity, err := repo.Find(ctx, id)
	logger.Info("TestReservationRepository_Find", "entity", entity)
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
