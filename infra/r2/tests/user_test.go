// R2 media repository tests. Require env vars (or a .env at repo root):
//
//	R2_BUCKET_NAME       - R2 bucket name
//	R2_ACCOUNT_ID        - Cloudflare account ID
//	R2_ACCESS_KEY_ID     - R2 API token access key
//	R2_ACCESS_KEY_SECRET - R2 API token secret key
//
// .env is loaded automatically from the repo root when present. Otherwise export the vars and run:
//
//	go test -v ./infra/r2/tests/ -run TestMediaRepository
package r2_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/user"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/r2"
	"github.com/gabriel-vasile/mimetype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
How to run the tests:
go test -v ./infra/r2/tests/ -run TestMediaRepository_Find
go test -v ./infra/r2/tests/ -run TestMediaRepository_Store
*/

func TestUserRepository_Store(t *testing.T) {
	client, bucketName := setupR2Client(t)
	repo := r2.NewUserR2Repository(client, bucketName)

	ctx := context.Background()
	id, err := user.NewUserID()
	require.NoError(t, err)

	content, err := os.ReadFile("saitama.webp")
	require.NoError(t, err)

	mt := mimetype.Detect(content)
	fmt.Println(mt.String())

	entity := &user.User{
		UserID:      id,
		Email:       "test@mail.com",
		Gender:      user.Gender_Male,
		CreatedTime: time.Now(),
		ImageSet: &user.ImageSet{
			ContentType: mt.String(),
			Content:     content,
		},
	}

	err = repo.Store(ctx, entity)
	fmt.Println(entity.ProfileKey)
	assert.NoError(t, err)
	assert.NotNil(t, entity.ProfileKey)
}

func TestUserRepository_Find(t *testing.T) {
	client, bucketName := setupR2Client(t)
	repo := r2.NewUserR2Repository(client, bucketName)

	ctx := context.Background()

	// Store an object first so we have a key to Find
	id, err := user.NewUserID()
	require.NoError(t, err)
	content, err := os.ReadFile("saitama.webp")
	require.NoError(t, err)
	mt := mimetype.Detect(content)
	entity := &user.User{
		UserID:      id,
		Email:       "test@mail.com",
		Gender:      user.Gender_Male,
		CreatedTime: time.Now(),
		ImageSet: &user.ImageSet{
			ContentType: mt.String(),
			Content:     content,
		},
	}
	err = repo.Store(ctx, entity)
	require.NoError(t, err)
	require.NotEmpty(t, entity.ProfileKey, "Store should set ProfileKey")

	// Find returns a presigned URL for the object
	url, err := repo.Find(ctx, *entity.ProfileKey)
	require.NoError(t, err)

	// Assert it looks like a URL (presigned URLs are https and have query params)
	assert.True(t, strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://"),
		"Find should return a URL: got %q", url)
	assert.Contains(t, url, "?", "presigned URL should contain query parameters")
	assert.Greater(t, len(url), 50, "presigned URL should be non-trivial")

	fmt.Println("Presigned URL:", url)
}
