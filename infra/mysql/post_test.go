package mysql_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/post"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/project"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/projectsvc"
	"github.com/stretchr/testify/assert"
)

func TestPostRepository_Find(t *testing.T) {
	db := mysql.SetupTestDB(t)

	repo := mysql.NewPostRepository(db)

	ctx := context.Background()

	id := post.PostID("c2a4fea6-7112-11f0-9198-8abfd21201dc")

	entity, err := repo.Find(ctx, id)
	err = mysql.PrettyPrint(entity)

	assert.NoError(t, err)
	assert.NotNil(t, entity)
	assert.Equal(t, id, entity.PostID)

}

func TestPostRepository_Store(t *testing.T) {
	db := mysql.SetupTestDB(t)

	repo := mysql.NewPostRepository(db)
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

	id, err := post.NewPostID()
	assert.NoError(t, err)

	postEntity := &post.Post{}
	postEntity.PostID = id
	postEntity.PostProjectID = post.ProjectID(projectEntity.ProjectID)
	content := "kantotent"
	postEntity.Content = &content

	visibility := post.Visibility_Public
	postEntity.Visibility = &visibility
	postEntity.CreatedTime = now
	postEntity.UpdatedTime = &now

	err = repo.Store(ctx, postEntity)
	assert.NoError(t, err)
}

func Test_PostRepository_Remove(t *testing.T) {
	db := mysql.SetupTestDB(t)

	repo := mysql.NewPostRepository(db)

	ctx := context.Background()

	err := repo.Remove(ctx, post.PostID("c2a4fea6-7112-11f0-9198-8abfd21201dc"))
	assert.NoError(t, err)
}

// SERVICE
func TestPostService_List(t *testing.T) {
	db := mysql.SetupTestDB(t)

	service := mysql.NewPostService(db)

	ctx := context.Background()

	// postID := post.PostID("c2a4fea6-7112-11f0-9198-8abfd21201dc")
	offset := int64(0)

	cond := post.ListCond{
		// PostID: &postID,
		SortKey: post.PostSortKey_CreatedTime_DESC,
		Offset:  &offset,
		Size:    10,
	}

	entities, err := service.List(ctx, cond)
	assert.NoError(t, err)
	assert.NotNil(t, entities)

	err = mysql.PrettyPrint(entities)
	assert.NoError(t, err)
}

func TestPostService_Count(t *testing.T) {
	db := mysql.SetupTestDB(t)

	service := mysql.NewPostService(db)

	ctx := context.Background()

	cond := post.CountCond{}
	count, err := service.Count(ctx, cond)
	assert.NoError(t, err)

	err = mysql.PrettyPrint(count)
	assert.NoError(t, err)
}

func TestPostService_ListPostStatsByProjectIds(t *testing.T) {
	db := mysql.SetupTestDB(t)

	service := mysql.NewPostService(db)

	ctx := context.Background()

	projectIds := []post.ProjectID{post.ProjectID("ac13a9cc-0f77-11f1-826f-8abfd21201dc")}

	cond := &post.ListPostStatsByProjectIdsCond{}
	out, err := service.ListPostStatsByProjectIds(ctx, *cond, projectIds...)
	assert.NoError(t, err)
	assert.NotNil(t, out)

	b, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println("dest:", string(b))

	err = mysql.PrettyPrint(out)
	assert.NoError(t, err)
}
