package main

import (
	"log"
	"video_processor/hlssegmenter"
	"video_processor/pubsub"

	"github.com/joho/godotenv"
)

func main() {
	fileName := "test.mp4"
	outputDir := "segments"

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Start the subscriber in a goroutine
	go pubsub.SubscribeToVideoProcessed()

	go hlssegmenter.StartSegmentProcess(fileName, outputDir)

	// clean up
	// err = utils.DeleteLocalFile(unprecessedVideoPath)
	// if err != nil {
	// 	fmt.Print(err.Error())
	// }

	// err = utils.DeleteDirContents(desireOutputPath)
	// if err != nil {
	// 	fmt.Print(err.Error())
	// }
	select {}
}
