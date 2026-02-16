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

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/media"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/infra/r2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gabriel-vasile/mimetype"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
How to run the tests:
go test -v ./infra/r2/tests/ -run TestMediaRepository_Find
go test -v ./infra/r2/tests/ -run TestMediaRepository_Store
*/

func init() {
	// Load .env so R2 tests work when run from IDE or without exporting vars.
	for _, path := range []string{".env", "../../.env"} {
		if _, err := os.Stat(path); err == nil {
			_ = godotenv.Load(path)
			break
		}
	}
}

// R2 env var names (set these to run the tests).
const (
	envR2BucketName  = "R2_BUCKET_NAME"
	envR2AccountID   = "R2_ACCOUNT_ID"
	envR2AccessKeyID = "R2_ACCESS_KEY_ID"
	envR2SecretKey   = "R2_ACCESS_KEY_SECRET"
)

// setupR2Client builds an S3 client for R2 using env vars. Skips t if env is not set.
func setupR2Client(t *testing.T) (*s3.Client, string) {
	t.Helper()

	fmt.Println("--------------------------------")
	fmt.Println(os.Getenv(envR2BucketName))
	fmt.Println(os.Getenv(envR2AccountID))
	fmt.Println(os.Getenv(envR2AccessKeyID))
	fmt.Println(os.Getenv(envR2SecretKey))
	fmt.Println("--------------------------------")

	bucketName := os.Getenv(envR2BucketName)
	accountID := os.Getenv(envR2AccountID)
	accessKeyID := os.Getenv(envR2AccessKeyID)
	secretKey := os.Getenv(envR2SecretKey)

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretKey,
			"",
		)),
		config.WithRegion("auto"),
	)
	require.NoError(t, err)
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("https://" + accountID + ".r2.cloudflarestorage.com")
	})
	return client, bucketName
}

func TestMediaRepository_Store(t *testing.T) {
	client, bucketName := setupR2Client(t)
	repo := r2.NewMediaR2Repository(client, bucketName)

	ctx := context.Background()
	id, err := media.NewMediaID()
	require.NoError(t, err)

	content, err := os.ReadFile("saitama.webp")
	require.NoError(t, err)

	mt := mimetype.Detect(content)
	fmt.Println(mt.String())

	entity := &media.Media{
		MediaID:     id,
		MediaPostID: "test-post-id",
		CreatedTime: time.Now(),
		MediaSet: media.MediaSet{
			ContentType: mt.String(),
			Content:     content,
		},
	}

	err = repo.Store(ctx, entity)
	fmt.Println(entity.ObjectKey)
	assert.NoError(t, err)
	assert.NotNil(t, entity.ObjectKey)
}

func TestMediaRepository_Find(t *testing.T) {
	client, bucketName := setupR2Client(t)
	repo := r2.NewMediaR2Repository(client, bucketName)

	ctx := context.Background()

	// Store an object first so we have a key to Find
	id, err := media.NewMediaID()
	require.NoError(t, err)
	content, err := os.ReadFile("saitama.webp")
	require.NoError(t, err)
	mt := mimetype.Detect(content)
	entity := &media.Media{
		MediaID:     id,
		MediaPostID: "test-post-id",
		CreatedTime: time.Now(),
		MediaSet: media.MediaSet{
			ContentType: mt.String(),
			Content:     content,
		},
	}
	err = repo.Store(ctx, entity)
	require.NoError(t, err)
	require.NotEmpty(t, entity.ObjectKey, "Store should set ObjectKey")

	// Find returns a presigned URL for the object
	url, err := repo.Find(ctx, entity.ObjectKey)
	require.NoError(t, err)

	// Assert it looks like a URL (presigned URLs are https and have query params)
	assert.True(t, strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://"),
		"Find should return a URL: got %q", url)
	assert.Contains(t, url, "?", "presigned URL should contain query parameters")
	assert.Greater(t, len(url), 50, "presigned URL should be non-trivial")

	fmt.Println("Presigned URL:", url)
}
