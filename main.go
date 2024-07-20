package main

import (
	"fmt"
	"sync"
	"video_processor/constants"
	"video_processor/hlssegmenter"
	"video_processor/resolutionparser"
	"video_processor/utils"
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

	execHLSSegmentVideo(outputDir)
	fmt.Println("All videos processed.")
}

func execHLSSegmentVideo(outputDir string) {
	videoNames, err := utils.GetVideoNames(outputDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, constants.MaxConcurrent)

	for _, name := range videoNames {
		wg.Add(1)
		sem <- struct{}{}

		go func(name string) {
			defer wg.Done()
			defer func() { <-sem }()
			hlssegmenter.Run(
				fmt.Sprintf("%s/%s", outputDir, name),
				fmt.Sprintf("%s/output_segs/%s", outputDir, utils.RemoveFileExtension(name)),
				"playlist.m3u8",
				5,
			)
		}(name)
	}

	wg.Wait()
}
