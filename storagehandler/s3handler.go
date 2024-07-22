package storagehandler

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// UploadFileToS3 uploads a file to an S3 bucket using the input file path as the S3 key
func UploadFileToS3(inputFilePath, bucketName, region string) error {
	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return fmt.Errorf("error creating AWS session: %w", err)
	}

	// Create an S3 service client
	svc := s3.New(sess)

	// Open the input file
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
		Key:    aws.String(fileName),
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
