package watermill

import (
	"encoding/json"
	"os"
	"video_processor/hlssegmenter"
	"video_processor/logger"
	"video_processor/messagemodel"

	"github.com/ThreeDotsLabs/watermill/message"
	"go.uber.org/zap"
)

func HandleNewVideoUploadEvent(msg *message.Message) {
	var videoInfo *messagemodel.VideoInfo
	err := json.Unmarshal(msg.Payload, &videoInfo)
	if err != nil {
		logger.AppLogger.Error("cannot unmarshal message", zap.Error(err), zap.Any("msg", msg))
		return
	}

	if videoInfo.RawVidS3Key == "" {
		logger.AppLogger.Error("s3key is empty", zap.Any("videoInfo", videoInfo))
		return
	}

	segmentOutputDir := os.Getenv("OUTPUT_SEGMENT_DIR")
	logcalOutputDir, err := hlssegmenter.StartSegmentProcess(videoInfo.RawVidS3Key, segmentOutputDir)

	if err != nil {
		logger.AppLogger.Error("cannot start segment process", zap.Error(err), zap.Any("S3Key", videoInfo.RawVidS3Key))
		return
	}

	processedSegmentsInfo := messagemodel.ProcessedSegmentsInfo{
		VideoSlug:      videoInfo.VideoSlug,
		CourseSlug:     videoInfo.CourseSlug,
		UserEmail:      videoInfo.UserEmail,
		LocalOutputDir: logcalOutputDir,
	}

	go VideoProcessedPublisher(processedSegmentsInfo)
	msg.Ack()
}
