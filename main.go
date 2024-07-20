package main

import (
	"fmt"
	"video_processor/hlssegmenter"
	"video_processor/resolutionparser"
)

func main() {
	fileName := "test.mp4"
	outputDir := "output"
	resolutions := []int{480, 720, 1080}

	resolutionparser.Run(
		fileName,
		fmt.Sprintf("%s/%s", outputDir, fileName),
		resolutions,
	)

	hlssegmenter.ExecHLSSegmentVideo(outputDir)
	fmt.Println("All videos processed.")
}
