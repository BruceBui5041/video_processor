package resolutionparser

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func Run(inputFile string, outputPrefix string, resolutions []int) {
	inputHeight, err := getVideoHeight(inputFile)
	if err != nil {
		log.Fatalf("Error getting input video height: %v", err)
	}

	log.Printf("Input video height: %d", inputHeight)

	for _, res := range resolutions {
		if res >= inputHeight {
			log.Printf("Skipping %dp resolution as it's higher than or equal to the input video height", res)
			continue
		}

		outputFile := fmt.Sprintf("%s_%dp.mp4", outputPrefix, res)
		err := segmentVideo(inputFile, outputFile, res)
		if err != nil {
			log.Printf("Error processing %dp resolution: %v", res, err)
		} else {
			log.Printf("Successfully created %dp segment", res)
		}
	}
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
