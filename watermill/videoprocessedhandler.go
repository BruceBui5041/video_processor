package watermill

import (
	"encoding/json"
	"video_processor/appconst"
	"video_processor/logger"
	"video_processor/messagemodel"
	"video_processor/storagehandler"
	"video_processor/utils"

	"github.com/ThreeDotsLabs/watermill/message"
	"go.uber.org/zap"
)

func HandleVideoProcessedVideoEvent(msg *message.Message) {
	var proccessedSegmentsInfo *messagemodel.ProcessedSegmentsInfo
	err := json.Unmarshal(msg.Payload, &proccessedSegmentsInfo)
	if err != nil {
		logger.AppLogger.Error(
			"cannot unmarchal msg payload",
			zap.String("payload",
				string(msg.Payload)),
			zap.Error(err),
		)
		msg.Ack()
		return
	}

	outputDir := proccessedSegmentsInfo.LocalOutputDir

	filePaths, err := utils.GetFilePaths(outputDir)
	if err != nil {
		logger.AppLogger.Error("Failed to get file paths", zap.Error(err), zap.String("outputDir", outputDir))
		msg.Ack()
		return
	}

	sem := make(chan struct{}, appconst.MaxConcurrentS3Push)
	for _, path := range filePaths {
		go func(path string) {
			sem <- struct{}{}
			defer func() { <-sem }()
			storagehandler.GenerateSegmentS3Key(storagehandler.VideoInfo{
				Useremail:  proccessedSegmentsInfo.UserEmail,
				CourseSlug: proccessedSegmentsInfo.CourseSlug,
				VideoSlug:  proccessedSegmentsInfo.VideoSlug,
				// Filename: ,
			})

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
