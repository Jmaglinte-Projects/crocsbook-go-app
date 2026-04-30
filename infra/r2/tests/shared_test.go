package r2_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func init() {
	loadR2TestEnv()
}

// R2 env var names (set these to run the tests).
const (
	envR2BucketName  = "R2_BUCKET_NAME"
	envR2AccountID   = "R2_ACCOUNT_ID"
	envR2AccessKeyID = "R2_ACCESS_KEY_ID"
	envR2SecretKey   = "R2_ACCESS_KEY_SECRET"
)

func loadR2TestEnv() {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return
	}

	testsDir := filepath.Dir(currentFile)
	repoRoot := filepath.Clean(filepath.Join(testsDir, "..", "..", ".."))

	candidates := []string{
		filepath.Join(testsDir, ".env"),
		filepath.Join(repoRoot, ".env"),
		filepath.Join(".", ".env"),
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			_ = godotenv.Load(path)
			return
		}
	}
}

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
