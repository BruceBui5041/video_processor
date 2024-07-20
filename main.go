package main

import (
	"fmt"
	"video_processor/resolutionparser"
)

func main() {
	fileName := "test"
	outputDir := "output"
	resolutions := []int{480, 720, 1080}
	resolutionparser.Run(
		fmt.Sprintf("%s.mp4", fileName),
		fmt.Sprintf("%s/%s", outputDir, fileName),
		resolutions,
	)
}
