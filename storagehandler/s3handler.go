package storagehandler

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"video_processor/appconst"
	"video_processor/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var awsConfig aws.Config

func init() {
	// Load the .env file
	if err := godotenv.Load(); err != nil {
		logger.AppLogger.Fatal("Error loading .env file", zap.Error(err))
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
		logger.AppLogger.Fatal("Unable to load SDK config", zap.Error(err))
	}

	awsConfig = cfg
	logger.AppLogger.Info("AWS configuration loaded successfully")
}

func UploadFileToS3(inputFilePath, bucketName string) error {
	client := s3.NewFromConfig(awsConfig)

	file, err := os.Open(inputFilePath)
	if err != nil {
		logger.AppLogger.Error("Error opening file", zap.Error(err), zap.String("filePath", inputFilePath))
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	fileName := filepath.Base(inputFilePath)

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(inputFilePath),
		Body:   file,
	})
	if err != nil {
		logger.AppLogger.Error("Error uploading file to S3", zap.Error(err), zap.String("bucket", bucketName), zap.String("key", inputFilePath))
		return fmt.Errorf("error uploading file to S3: %w", err)
	}

	logger.AppLogger.Info("File uploaded successfully", zap.String("filePath", inputFilePath), zap.String("bucket", bucketName), zap.String("key", fileName))
	return nil
}

func GetS3File(bucket, key, saveDir string) (string, error) {
	client := s3.NewFromConfig(awsConfig)

	result, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		logger.AppLogger.Error("Failed to get object from S3", zap.Error(err), zap.String("bucket", bucket), zap.String("key", key))
		return "", fmt.Errorf("failed to get object: %v", err)
	}
	defer result.Body.Close()

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		logger.AppLogger.Error("Failed to create save directory", zap.Error(err), zap.String("directory", saveDir))
		return "", fmt.Errorf("failed to create save directory: %v", err)
	}

	fileName := filepath.Base(key)
	localPath := filepath.Join(saveDir, fileName)

	file, err := os.Create(localPath)
	if err != nil {
		logger.AppLogger.Error("Failed to create local file", zap.Error(err), zap.String("filePath", localPath))
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, result.Body)
	if err != nil {
		logger.AppLogger.Error("Failed to copy content from S3 to local file", zap.Error(err), zap.String("filePath", localPath))
		return "", fmt.Errorf("failed to copy content: %v", err)
	}

	logger.AppLogger.Info("File downloaded successfully from S3", zap.String("bucket", bucket), zap.String("key", key), zap.String("localPath", localPath))
	return localPath, nil
}
