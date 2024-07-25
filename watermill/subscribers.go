package watermill

import (
	"context"
	"fmt"
	"video_processor/appconst"
	"video_processor/logger"

	"go.uber.org/zap"
)

func SubscribeToTopics() {
	subscriber := Publisher

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create channels for each topic
	videoProcessedChan, err := subscriber.Subscribe(ctx, appconst.TopicVideoProcessed)
	if err != nil {
		logger.AppLogger.Fatal(fmt.Sprintf("Failed to subscribe to %s topic", appconst.TopicVideoProcessed), zap.Error(err))
	}

	newVideoUploadedChan, err := subscriber.Subscribe(ctx, appconst.TopicNewVideoUploaded)
	if err != nil {
		logger.AppLogger.Fatal(fmt.Sprintf("Failed to subscribe to %s topic", appconst.TopicNewVideoUploaded), zap.Error(err))
	}

	for {
		select {
		case msg := <-videoProcessedChan:
			go HandleVideoProcessedVideoEvent(msg)
		case msg := <-newVideoUploadedChan:
			go HandleNewVideoUploadEvent(msg)
		case <-ctx.Done():
			return
		}
	}
}
