package hlssegmenter

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func Run(inputFile string, outputDir string, playlistName string, segmentDuration int) {
	// Check if FFmpeg is installed
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		log.Fatal("FFmpeg not found. Please install FFmpeg to continue.")
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Get video duration
	duration, err := getVideoDuration(inputFile)
	if err != nil {
		log.Fatalf("Failed to get video duration: %v", err)
	}

	// Generate the FFmpeg command for HLS segmentation
	cmd := generateFFmpegCommand(inputFile, outputDir, playlistName, segmentDuration)

	// Create a pipe to capture FFmpeg output
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("Failed to create stderr pipe: %v", err)
	}

	// Start the FFmpeg command
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start FFmpeg: %v", err)
	}

	// Start a goroutine to read FFmpeg output and update progress
	go func() {
		scanner := bufio.NewScanner(stderr)
		re := regexp.MustCompile(`time=(\d{2}):(\d{2}):(\d{2})\.(\d{2})`)
		for scanner.Scan() {
			line := scanner.Text()
			matches := re.FindStringSubmatch(line)
			if len(matches) == 5 {
				hours, _ := strconv.Atoi(matches[1])
				minutes, _ := strconv.Atoi(matches[2])
				seconds, _ := strconv.Atoi(matches[3])
				processedDuration := time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second
				progress := float64(processedDuration) / float64(duration) * 100
				fmt.Printf("\rProgress: %.2f%%", progress)
			}
		}
	}()

	// Wait for FFmpeg to finish
	if err := cmd.Wait(); err != nil {
		log.Fatalf("FFmpeg command failed: %v", err)
	}

	fmt.Println("\nHLS segmentation completed successfully.")
	fmt.Printf("Output files are in the '%s' directory.\n", outputDir)
}

// getVideoDuration returns the duration of the input video file
func getVideoDuration(inputFile string) (time.Duration, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", inputFile)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	durationStr := strings.TrimSpace(string(output))
	durationSec, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, err
	}
	return time.Duration(durationSec * float64(time.Second)), nil
}

// generateFFmpegCommand creates the FFmpeg command for HLS segmentation
func generateFFmpegCommand(inputFile string, outputDir string, playlistName string, segmentDuration int) *exec.Cmd {
	outputPath := filepath.Join(outputDir, "segment_%03d.ts")
	playlistPath := filepath.Join(outputDir, playlistName)

	// Construct the FFmpeg command
	args := []string{
		"-i", inputFile, // Input file
		"-profile:v", "baseline", // Use baseline profile for better device compatibility
		"-level", "3.0", // Set H.264 level
		"-start_number", "0", // Start segment numbering at 0
		"-hls_time", fmt.Sprintf("%d", segmentDuration), // Set segment duration
		"-hls_list_size", "0", // Keep all segments in the playlist
		"-f", "hls", // Force HLS output format
		"-hls_segment_filename", outputPath, // Set output segment file pattern
		playlistPath, // Set the output playlist file
	}

	// Create the command
	cmd := exec.Command("ffmpeg", args...)

	// Print the command for debugging
	fmt.Println("Executing FFmpeg command:")
	fmt.Println(strings.Join(cmd.Args, " "))

	return cmd
}
