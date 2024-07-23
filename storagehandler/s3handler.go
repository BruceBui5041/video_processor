package storagehandler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"video_processor/constants"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var AWSSession *session.Session

func init() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(constants.AWSRegion),
	})
	if err != nil {
		panic(err)
	}

	AWSSession = sess
}

func UploadFileToS3(inputFilePath, bucketName string) error {
	svc := s3.New(AWSSession)

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

func GetS3File(bucket, key, saveDir string) (string, error) {
	// Create an S3 service client
	svc := s3.New(AWSSession)

	// Create the GetObject request
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	// Fetch the object
	result, err := svc.GetObject(input)
	if err != nil {
		return "", fmt.Errorf("failed to get object: %v", err)
	}
	defer result.Body.Close()

	// Ensure the save directory exists
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create save directory: %v", err)
	}

	// Create the full path for the local file
	fileName := filepath.Base(key)
	localPath := filepath.Join(saveDir, fileName)

	// Create a local file to save the S3 object content
	file, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Copy the S3 object content to the local file
	_, err = io.Copy(file, result.Body)
	if err != nil {
		return "", fmt.Errorf("failed to copy content: %v", err)
	}

	fmt.Printf("Successfully downloaded %s from bucket %s to %s\n", key, bucket, localPath)
	return localPath, nil
}
