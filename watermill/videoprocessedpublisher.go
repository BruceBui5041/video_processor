package watermill

import (
	"encoding/json"
	"video_processor/appconst"
	"video_processor/logger"
	"video_processor/messagemodel"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"go.uber.org/zap"
)

func VideoProcessedPublisher(segmentsInfo messagemodel.ProcessedSegmentsInfo) {
	data, err := json.Marshal(segmentsInfo)
	if err != nil {
		logger.AppLogger.Error("cannot marshal", zap.Error(err))
		return
	}

	msg := message.NewMessage(watermill.NewUUID(), data)
	if err := Publisher.Publish(appconst.TopicVideoProcessed, msg); err != nil {
		logger.AppLogger.Error("Failed to publish video_processed event", zap.Error(err))
	}
}
