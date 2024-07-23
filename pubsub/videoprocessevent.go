package pubsub

import (
	"context"
	"fmt"
	"log"
	"strings"
	"video_processor/constants"
	"video_processor/storagehandler"
	"video_processor/utils"

	"github.com/ThreeDotsLabs/watermill/message"
)

func SubscribeToVideoProcessed() {
	subscriber := Publisher

	messages, err := subscriber.Subscribe(context.Background(), constants.TopicVideoProcessed)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	for msg := range messages {
		processMessage(msg)
	}
}

func processMessage(msg *message.Message) {
	parts := strings.Split(string(msg.Payload), ",")
	if len(parts) != 2 {
		log.Printf("Invalid message format: %s", msg.Payload)
		return
	}

	inputFile, outputDir := parts[0], parts[1]
	fmt.Printf("Video processed event received: input file: %s, output directory: %s\n", inputFile, outputDir)

	filePaths, err := utils.GetFilePaths(outputDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	sem := make(chan struct{}, constants.MaxConcurrentS3Push)
	for _, path := range filePaths {
		go func(path string) {
			sem <- struct{}{}
			defer func() { <-sem }()

			err := storagehandler.UploadFileToS3(path, constants.AWSVideoS3BuckerName, constants.AWSRegion)
			if err != nil {
				log.Print(err.Error())
			}
		}(path)
	}

	// Mark the message as processed
	msg.Ack()
}
