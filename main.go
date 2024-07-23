package main

import (
	"log"
	"path/filepath"
	"video_processor/constants"
	"video_processor/hlssegmenter"
	"video_processor/pubsub"
	"video_processor/storagehandler"
	"video_processor/utils"
)

func main() {
	fileName := "test.mp4"
	outputDir := "segments"
	unprecessedVideoDir := "unprocessed_video"

	// Start the subscriber in a goroutine
	go pubsub.SubscribeToVideoProcessed()

	utils.CreateDirIfNotExist(unprecessedVideoDir)
	unprecessedVideoPath, err := storagehandler.GetS3File(
		constants.AWSVideoS3BuckerName,
		fileName,
		unprecessedVideoDir,
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	desireOutputPath := filepath.Join(outputDir, fileName)
	utils.CreateDirIfNotExist(desireOutputPath)
	hlssegmenter.ExecHLSSegmentVideo(unprecessedVideoPath, desireOutputPath)

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
