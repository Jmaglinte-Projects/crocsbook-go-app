package r2

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/project"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/projectsvc"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type projectR2Repository struct {
	r2         *s3.Client
	bucketName string
}

func NewProjectR2Repository(r2 *s3.Client, bucketName string) projectsvc.ProjectR2Repository {
	return &projectR2Repository{
		r2:         r2,
		bucketName: bucketName,
	}
}

func (r *projectR2Repository) Find(ctx context.Context, key string) (string, error) {
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

func (r *projectR2Repository) Store(ctx context.Context, entity *project.Project) error {
	if len(entity.ThumbnailSet.Content) == 0 {
		fmt.Println("No image bytes found to upload to r2")
		return nil
	}

	if entity.ThumbnailSet.ContentType == "" {
		entity.ThumbnailSet.ContentType = "application/octet-stream"
	}

	now := time.Now()
	year := now.Year()
	month := now.Month()
	day := now.Day()

	fmt.Println("---------BUCKET NAME------------")
	fmt.Println(r.bucketName)
	fmt.Println("--------------------------------")

	// project/2026/January/01/XXXXXXX
	key := fmt.Sprintf("project/%d/%s/%d/%s", year, month.String(), day, entity.ProjectID)
	putObjectInput := &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(entity.ThumbnailSet.Content),
		ContentType: aws.String(entity.ThumbnailSet.ContentType),
	}

	fmt.Println("--------------------------------")
	fmt.Println("key: ", key)
	fmt.Println("--------------------------------")

	_, err := r.r2.PutObject(ctx, putObjectInput)
	if err != nil {
		fmt.Println("Error uploading image to r2")
		return err
	}

	entity.Thumbnail = &key

	return nil
}

func (r *projectR2Repository) Remove(ctx context.Context, keys ...string) error {

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

// TODO: Implement the project repository
