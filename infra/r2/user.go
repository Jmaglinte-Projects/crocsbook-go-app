package r2

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/user"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/usersvc"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

/*
# How to test the userR2Repository:

go test -v ./infra/r2/tests/ -run TestUserRepository

# or specific ones:
go test -v ./infra/r2/tests/ -run TestUserRepository_Store
go test -v ./infra/r2/tests/ -run TestUserRepository_Find
*/

type userR2Repository struct {
	r2         *s3.Client
	bucketName string
}

func NewUserR2Repository(r2 *s3.Client, bucketName string) usersvc.UserR2Repository {
	return &userR2Repository{
		r2:         r2,
		bucketName: bucketName,
	}
}

func (r *userR2Repository) Find(ctx context.Context, key string) (string, error) {
	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	}

	presignClient := s3.NewPresignClient(r.r2)
	presigned, err := presignClient.PresignGetObject(ctx, getObjectInput, s3.WithPresignExpires(time.Hour))
	if err != nil {
		return "", err
	}
	return presigned.URL, nil
}

func (r *userR2Repository) Store(ctx context.Context, entity *user.User) error {
	if len(entity.ImageSet.Content) == 0 {
		fmt.Println("No image bytes found to upload to r2")
		return nil
	}

	if entity.ImageSet.ContentType == "" {
		entity.ImageSet.ContentType = "application/octet-stream"
	}

	now := time.Now()
	year := now.Year()
	month := now.Month()
	day := now.Day()

	fmt.Println("---------BUCKET NAME------------")
	fmt.Println(r.bucketName)
	fmt.Println("--------------------------------")

	// user/2026/January/01/:id
	key := fmt.Sprintf("user/%d/%s/%d/%s", year, month.String(), day, entity.UserID)
	putObjectInput := &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(entity.ImageSet.Content),
		ContentType: aws.String(entity.ImageSet.ContentType),
	}

	fmt.Println("--------------------------------")
	fmt.Println("key: ", key)
	fmt.Println("--------------------------------")

	_, err := r.r2.PutObject(ctx, putObjectInput)
	if err != nil {
		fmt.Println("Error uploading image to r2")
		return err
	}

	entity.ProfileKey = &key

	return nil
}

func (r *userR2Repository) Remove(ctx context.Context, keys ...string) error {

	for _, key := range keys {
		deleteObjectInput := &s3.DeleteObjectInput{
			Bucket: aws.String(r.bucketName),
			Key:    aws.String(key),
		}
		_, err := r.r2.DeleteObject(ctx, deleteObjectInput)
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: Implement the media repository
