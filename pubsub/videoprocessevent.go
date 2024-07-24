package pubsub

import (
	"context"
	"strings"
	"video_processor/appconst"
	"video_processor/logger"
	"video_processor/storagehandler"
	"video_processor/utils"

	"github.com/ThreeDotsLabs/watermill/message"
	"go.uber.org/zap"
)

func SubscribeToVideoProcessed() {
	subscriber := Publisher

	messages, err := subscriber.Subscribe(context.Background(), appconst.TopicVideoProcessed)
	if err != nil {
		logger.AppLogger.Fatal("Failed to subscribe", zap.Error(err))
	}

	for msg := range messages {
		processMessage(msg)
	}
}

func processMessage(msg *message.Message) {
	parts := strings.Split(string(msg.Payload), ",")
	if len(parts) != 2 {
		logger.AppLogger.Error("Invalid message format", zap.String("payload", string(msg.Payload)))
		return
	}

	inputFile, outputDir := parts[0], parts[1]
	logger.AppLogger.Info("Video processed event received",
		zap.String("inputFile", inputFile),
		zap.String("outputDir", outputDir))

	filePaths, err := utils.GetFilePaths(outputDir)
	if err != nil {
		logger.AppLogger.Error("Failed to get file paths", zap.Error(err), zap.String("outputDir", outputDir))
		return
	}

	sem := make(chan struct{}, appconst.MaxConcurrentS3Push)
	for _, path := range filePaths {
		go func(path string) {
			sem <- struct{}{}
			defer func() { <-sem }()

			err := storagehandler.UploadFileToS3(path, appconst.AWSVideoS3BuckerName)
			if err != nil {
				logger.AppLogger.Error("Failed to upload file to S3",
					zap.Error(err),
					zap.String("path", path),
					zap.String("bucket", appconst.AWSVideoS3BuckerName))
			} else {
				logger.AppLogger.Info("File uploaded to S3 successfully",
					zap.String("path", path),
					zap.String("bucket", appconst.AWSVideoS3BuckerName))
			}
		}(path)
	}

	// Mark the message as processed
	msg.Ack()
	logger.AppLogger.Info("Message processed and acknowledged", zap.String("messageID", msg.UUID))
}
