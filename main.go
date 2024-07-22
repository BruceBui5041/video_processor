package main

import (
	"fmt"
	"video_processor/hlssegmenter"
)

func main() {
	fileName := "test.mp4"
	outputDir := "output"
	// resolutions := []int{480, 720, 1080}

	// resolutionparser.Run(
	// 	fileName,
	// 	fmt.Sprintf("%s/%s", outputDir, fileName),
	// 	resolutions,
	// )

	hlssegmenter.ExecHLSSegmentVideo(fileName, outputDir)
	fmt.Println("All videos processed.")
}
