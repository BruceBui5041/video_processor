package storagehandler

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"video_processor/appconst"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

var awsConfig aws.Config

func init() {
	// Load the .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get credentials from environment variables
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	// Create a new credential provider
	creds := credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")

	// Load the configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(appconst.AWSRegion),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	awsConfig = cfg
}

func UploadFileToS3(inputFilePath, bucketName string) error {
	client := s3.NewFromConfig(awsConfig)

	file, err := os.Open(inputFilePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Get the file name from the input path
	fileName := filepath.Base(inputFilePath)

	// Upload the file to S3
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(inputFilePath),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("error uploading file to S3: %w", err)
	}

	fmt.Printf("Successfully uploaded %s to s3://%s/%s\n", inputFilePath, bucketName, fileName)
	return nil
}

func GetS3File(bucket, key, saveDir string) (string, error) {
	client := s3.NewFromConfig(awsConfig)

	// Fetch the object
	result, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
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
