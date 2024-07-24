package resolutionparser

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"video_processor/appconst"
	"video_processor/logger"

	"go.uber.org/zap"
)

func Run(inputFile string, outputPrefix string, resolutions []int) {
	inputHeight, err := getVideoHeight(inputFile)
	if err != nil {
		logger.AppLogger.Fatal("Error getting input video height", zap.Error(err))
	}

	logger.AppLogger.Info("Input video height", zap.Int("height", inputHeight))

	var wg sync.WaitGroup
	sem := make(chan struct{}, appconst.VideoMaxConcurrentResolutionParse)

	for _, res := range resolutions {
		if res >= inputHeight {
			logger.AppLogger.Info("Skipping resolution", zap.Int("resolution", res), zap.String("reason", "higher than or equal to input video height"))
			continue
		}

		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore
		go func(res int) {
			defer wg.Done()
			defer func() { <-sem }() // Release the semaphore when done

			outputFile := fmt.Sprintf("%s_%dp.mp4", outputPrefix, res)
			err := segmentVideo(inputFile, outputFile, res)
			if err != nil {
				logger.AppLogger.Error("Error processing resolution", zap.Int("resolution", res), zap.Error(err))
			} else {
				logger.AppLogger.Info("Successfully created segment", zap.Int("resolution", res))
			}
		}(res)
	}

	wg.Wait()
}

func getVideoHeight(input string) (int, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-count_packets", "-show_entries", "stream=height",
		"-of", "csv=p=0",
		input,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %v", err)
	}

	height, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, fmt.Errorf("failed to parse video height: %v", err)
	}

	return height, nil
}

func segmentVideo(input string, output string, resolution int) error {
	cmd := exec.Command("ffmpeg",
		"-i", input,
		"-vf", fmt.Sprintf("scale=-2:%d", resolution),
		"-c:v", "libx264",
		"-crf", "23",
		"-preset", "medium",
		"-c:a", "aac",
		"-b:a", "128k",
		"-movflags", "+faststart",
		"-y",
		output,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
