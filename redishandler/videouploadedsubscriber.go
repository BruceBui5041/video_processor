package redishander

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"video_processor/appconst"
	"video_processor/logger"
	"video_processor/watermill"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func StartRedisSubscribers(redisClient *redis.Client) {
	ctx := context.Background()
	pubsub := redisClient.Subscribe(ctx, appconst.TopicNewVideoUploaded)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for msg := range ch {
		log.Printf("Received message from Redis channel %s: %s", msg.Channel, msg.Payload)

		// Parse the message payload
		var videoInfo struct {
			VideoID string `json:"video_id"`
			// Add other fields as needed
		}
		err := json.Unmarshal([]byte(msg.Payload), &videoInfo)
		if err != nil {
			log.Printf("Error parsing message payload: %v", err)
			continue
		}

		// Create a Watermill message
		watermillMsg := message.NewMessage(videoInfo.VideoID, []byte(msg.Payload))

		// Process the message using the existing handler
		if err := watermill.Publisher.Publish(appconst.TopicNewVideoUploaded, watermillMsg); err != nil {
			logger.AppLogger.Error(fmt.Sprintf("Failed to publish %s event", appconst.TopicNewVideoUploaded), zap.Error(err))
		}
	}
}
