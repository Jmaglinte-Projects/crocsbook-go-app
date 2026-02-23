package mysql_test

import (
	"context"
	"testing"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/media"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/project"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/mediasvc"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/projectsvc"
	"github.com/stretchr/testify/assert"
)

func TestMediaRepository_Find(t *testing.T) {
	db := mysql.SetupTestDB(t)

	repo := mysql.NewMediaRepository(db, nil)

	ctx := context.Background()

	id := media.MediaID("81783e14-0f7f-11f1-826f-8abfd21201dc")

	entity, err := repo.Find(ctx, id)
	err = mysql.PrettyPrint(entity)

	assert.NoError(t, err)
	assert.NotNil(t, entity)
	assert.Equal(t, id, entity.MediaID)
}

func TestMediaRepository_Store(t *testing.T) {
	db := mysql.SetupTestDB(t)

	repo := mysql.NewMediaRepository(db, nil)
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

	id, err := media.NewMediaID()
	assert.NoError(t, err)

	mediaEntity := &media.Media{}
	mediaEntity.MediaID = id
	mediaEntity.MediaPostID = media.PostID(projectEntity.ProjectID)

	mediaEntity.Type = media.Type("image/webp")
	mediaEntity.CreatedTime = now

	err = repo.Store(ctx, mediaEntity)
	assert.NoError(t, err)
}

func Test_MediaRepository_Remove(t *testing.T) {
	db := mysql.SetupTestDB(t)

	repo := mysql.NewMediaRepository(db, nil)

	ctx := context.Background()

	err := repo.Remove(ctx, media.MediaID("887a2752-9911-11f0-afd3-8abfd21201dc"))
	assert.NoError(t, err)
}

// SERVICE
func TestMediaService_List(t *testing.T) {
	db := mysql.SetupTestDB(t)

	service := mysql.NewMediaService(db, nil)

	ctx := context.Background()

	// mediaID := media.MediaID("887a2752-9911-11f0-afd3-8abfd21201dc")
	offset := int64(0)

	cond := media.ListCond{
		// MediaID: &mediaID,
	}

	opt := mediasvc.ListOption{
		SortKey: mediasvc.ListOptionSortKey_CreatedAt_ASC,
		Offset:  &offset,
		Size:    10,
	}

	entities, err := service.List(ctx, cond, opt)
	assert.NoError(t, err)
	assert.NotNil(t, entities)

	err = mysql.PrettyPrint(entities)
	assert.NoError(t, err)
}

func TestMediaService_Count(t *testing.T) {
	db := mysql.SetupTestDB(t)

	service := mysql.NewMediaService(db, nil)

	ctx := context.Background()

	// mediaID := media.MediaID("887a2752-9911-11f0-afd3-8abfd21201dcx")
	cond := media.CountCond{
		// MediaID: &mediaID,
	}
	opt := mediasvc.CountOption{}

	count, err := service.Count(ctx, cond, opt)
	assert.NoError(t, err)

	err = mysql.PrettyPrint(count)
	assert.NoError(t, err)
}
