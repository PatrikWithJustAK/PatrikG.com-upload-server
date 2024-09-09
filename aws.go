package main

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var s3Client *s3.Client
var imglist []string

func init() {
	// Load the AWS configuration
	// AWS config is out of scope for this project, refer to AWS SDK docs
	// I have opted to use an IAM role with specific bucket permissions
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
		log.Fatalf("Check your ~/.aws/credentials")
	}

	// Create an S3 service client
	s3Client = s3.NewFromConfig(cfg)
}

func uploadFileToS3(file multipart.File, fileHeader *multipart.FileHeader, bucketName string) (string, error) {
	defer file.Close()

	// Generate a unique file name, timestamp + filepath
	fileName := fmt.Sprintf("%d-%s", time.Now().Unix(), filepath.Base(fileHeader.Filename))

	// Upload the file to S3
	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(fileName),
		Body:        file,
		ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
		//TODO: if we add a key/value pair for "isPublic", we can display a checkbox that determines if the file should be visible publicly or only from client
		// May also add tags to allow for sorting of filetypes
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %v", err)
	}

	// Return the URL of the uploaded file
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, fileName), nil
}

func getObjectURL(bucket, key string) string {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, url.PathEscape(key))
}

func listS3Objects(bucket string) []string {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	maxfiles := int32(9)
	// Create a ListObjectsV2 API request
	req := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		MaxKeys: &maxfiles,
	}

	// List all objects in the specified bucket
	resp, err := s3Client.ListObjectsV2(context.TODO(), req)
	if err != nil {
		log.Fatalf("unable to list items in bucket %q, %v", bucket, err)
	}

	// grab the URL for each object in the bucket, append to the imglist
	for _, item := range resp.Contents {
		objectURL := getObjectURL(bucket, *item.Key)
		imglist = append(imglist, objectURL)
	}
	return imglist
}
