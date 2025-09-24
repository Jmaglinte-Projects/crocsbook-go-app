package mysql_test

import (
	"context"
	"testing"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/user"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/mysql"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/usersvc"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_Find(t *testing.T) {
	db := mysql.SetupTestDB(t)

	repo := mysql.NewUserRepository(db)

	ctx := context.Background()

	id := user.UserID("c2a4fea6-7112-11f0-9198-8abfd21201dc")

	entity, err := repo.Find(ctx, id)
	err = mysql.PrettyPrint(entity)

	assert.NoError(t, err)
	assert.NotNil(t, entity)
	assert.Equal(t, id, entity.UserID)

}

func TestUserRepository_Store(t *testing.T) {
	db := mysql.SetupTestDB(t)

	repo := mysql.NewUserRepository(db)

	ctx := context.Background()
	now := time.Now()

	id := user.UserID("c2a4fea6-7112-11f0-9198-8abfd21201dc")

	nickName := "gwapoKo"
	entity := &user.User{
		UserID:      id,
		Email:       "test@mail.com",
		Gender:      "Male",
		ProfileURL:  nil,
		Nickname:    &nickName,
		CreatedTime: now,
		UpdatedTime: &now,
	}

	err := repo.Store(ctx, entity)
	assert.NoError(t, err)
}

// SERVICE
func TestUserService_List(t *testing.T) {
	db := mysql.SetupTestDB(t)

	service := mysql.NewUserService(db)

	ctx := context.Background()

	userID := user.UserID("c2a4fea6-7112-11f0-9198-8abfd21201dc")
	offset := int64(0)

	cond := user.ListCond{
		UserID: &userID,
	}

	opt := usersvc.ListOption{
		SortKey: usersvc.ListOptionSortKey_CreatedAt_ASC,
		Offset:  &offset,
		Size:    10,
	}

	entities, err := service.List(ctx, cond, opt)
	assert.NoError(t, err)
	assert.NotNil(t, entities)

	err = mysql.PrettyPrint(entities)
	assert.NoError(t, err)
}

func TestUserService_Count(t *testing.T) {
	db := mysql.SetupTestDB(t)

	service := mysql.NewUserService(db)

	ctx := context.Background()

	cond := user.CountCond{}
	opt := usersvc.CountOption{}

	count, err := service.Count(ctx, cond, opt)
	assert.NoError(t, err)

	err = mysql.PrettyPrint(count)
	assert.NoError(t, err)
}
