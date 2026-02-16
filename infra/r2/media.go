package r2

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/Jmaglinte-Projects/crocsbook-go-app/domain/media"
	"github.com/Jmaglinte-Projects/crocsbook-go-app/usecase/mediasvc"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type mediaR2Repository struct {
	r2         *s3.Client
	bucketName string
}

func NewMediaR2Repository(r2 *s3.Client, bucketName string) mediasvc.MediaR2Repository {
	return &mediaR2Repository{
		r2:         r2,
		bucketName: bucketName,
	}
}

func (r *mediaR2Repository) Find(ctx context.Context, key string) (string, error) {
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

func (r *mediaR2Repository) Store(ctx context.Context, entity *media.Media) error {
	if len(entity.MediaSet.Content) == 0 {
		fmt.Println("No image bytes found to upload to r2")
		return nil
	}

	if entity.MediaSet.ContentType == "" {
		entity.MediaSet.ContentType = "application/octet-stream"
	}

	now := time.Now()
	year := now.Year()
	month := now.Month()
	day := now.Day()

	fmt.Println("---------BUCKET NAME------------")
	fmt.Println(r.bucketName)
	fmt.Println("--------------------------------")

	// media/2026/January/01/XXXXXXX
	key := fmt.Sprintf("media/%d/%s/%d/%s", year, month.String(), day, entity.MediaID)
	putObjectInput := &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(entity.MediaSet.Content),
		ContentType: aws.String(entity.MediaSet.ContentType),
	}

	fmt.Println("--------------------------------")
	fmt.Println("key: ", key)
	fmt.Println("--------------------------------")

	_, err := r.r2.PutObject(ctx, putObjectInput)
	if err != nil {
		fmt.Println("Error uploading image to r2")
		return err
	}

	entity.ObjectKey = key

	return nil
}

func (r *mediaR2Repository) Remove(ctx context.Context, keys ...string) error {

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
