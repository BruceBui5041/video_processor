package resolutionparser

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func Run(inputFile string, outputPrefix string, resolutions []string) {
	for _, res := range resolutions {
		outputFile := fmt.Sprintf("%s_%sp.mp4", outputPrefix, res)
		err := segmentVideo(inputFile, outputFile, res)
		if err != nil {
			log.Printf("Error processing %sp resolution: %v", res, err)
		} else {
			log.Printf("Successfully created %sp segment", res)
		}
	}
}

func segmentVideo(input, output, resolution string) error {
	cmd := exec.Command("ffmpeg",
		"-i", input,
		"-vf", fmt.Sprintf("scale=-2:%s", resolution),
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
