package storagehandler

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func UploadFileToS3(inputFilePath, bucketName, region string) error {
	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return fmt.Errorf("error creating AWS session: %w", err)
	}

	svc := s3.New(sess)

	file, err := os.Open(inputFilePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Get the file name from the input path
	fileName := filepath.Base(inputFilePath)

	// Create the input for PutObject
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(inputFilePath),
		Body:   file,
	}

	// Upload the file to S3
	_, err = svc.PutObject(input)
	if err != nil {
		return fmt.Errorf("error uploading file to S3: %w", err)
	}

	fmt.Printf("Successfully uploaded %s to s3://%s/%s\n", inputFilePath, bucketName, fileName)
	return nil
}
