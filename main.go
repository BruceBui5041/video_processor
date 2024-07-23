package main

import (
	"fmt"
	"path/filepath"
	"video_processor/hlssegmenter"
	"video_processor/pubsub"
	"video_processor/utils"
)

func main() {
	fileName := "test.mp4"
	outputDir := "segments"

	// Start the subscriber in a goroutine
	go pubsub.SubscribeToVideoProcessed()

	desireOutputPath := filepath.Join(outputDir, fileName)
	utils.CreateDirIfNotExist(desireOutputPath)
	hlssegmenter.ExecHLSSegmentVideo(fileName, desireOutputPath)

	fmt.Println("All videos processed.")

	select {}
}
