package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Constants for HLS segmentation
const (
	segmentDuration = 10 // Duration of each segment in seconds
	masterPlaylist  = "playlist.m3u8"
)

// Resolution represents a video resolution with its corresponding bitrate
type Resolution struct {
	Name    string
	Width   int
	Height  int
	Bitrate string
}

// Available resolutions
var resolutions = []Resolution{
	{Name: "480p", Width: 854, Height: 480, Bitrate: "1000k"},
	{Name: "720p", Width: 1280, Height: 720, Bitrate: "2500k"},
	{Name: "1080p", Width: 1920, Height: 1080, Bitrate: "5000k"},
}

func main() {
	// Check if FFmpeg is installed
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		log.Fatal("FFmpeg not found. Please install FFmpeg to continue.")
	}

	// Get the input video file from command line arguments
	if len(os.Args) < 2 {
		log.Fatal("Please provide the input video file path as an argument.")
	}
	inputFile := os.Args[1]

	// Get video duration
	duration, err := getVideoDuration(inputFile)
	if err != nil {
		log.Fatalf("Failed to get video duration: %v", err)
	}

	// Generate the FFmpeg command for multi-resolution HLS segmentation
	cmd := generateFFmpegCommand(inputFile)

	// Create pipes for stdout and stderr
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	// Start the FFmpeg command
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start FFmpeg: %v", err)
	}

	// Create a buffer to store the error output
	var errorBuffer bytes.Buffer

	// Start goroutines to handle stdout and stderr
	go io.Copy(&errorBuffer, stdout)
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
			errorBuffer.WriteString(line + "\n")
		}
	}()

	// Wait for FFmpeg to finish
	if err := cmd.Wait(); err != nil {
		log.Printf("FFmpeg command failed: %v", err)
		log.Printf("FFmpeg error output:\n%s", errorBuffer.String())
		os.Exit(1)
	}

	fmt.Println("\nHLS segmentation completed successfully.")
	fmt.Println("Output files are in the 'output' directory.")
}

// getVideoDuration returns the duration of the input video file
func getVideoDuration(inputFile string) (time.Duration, error) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", inputFile)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %v", err)
	}

	var result struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return 0, fmt.Errorf("failed to parse ffprobe output: %v", err)
	}

	durationSec, err := strconv.ParseFloat(result.Format.Duration, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %v", err)
	}

	return time.Duration(durationSec * float64(time.Second)), nil
}

// generateFFmpegCommand creates the FFmpeg command for multi-resolution HLS segmentation
func generateFFmpegCommand(inputFile string) *exec.Cmd {
	args := []string{
		"-i", inputFile,
		"-keyint_min", "48",
		"-g", "48",
		"-sc_threshold", "0",
		"-r", "24",
		"-c:a", "aac",
		"-b:a", "128k",
		"-ar", "48000",
		"-ac", "2",
	}

	for _, res := range resolutions {
		args = append(args,
			"-filter:v:"+res.Name, "scale=w="+strconv.Itoa(res.Width)+":h="+strconv.Itoa(res.Height)+":force_original_aspect_ratio=decrease",
			"-c:v:"+res.Name, "libx264",
			"-b:v:"+res.Name, res.Bitrate,
			"-maxrate:v:"+res.Name, res.Bitrate,
			"-bufsize:v:"+res.Name, res.Bitrate,
			"-preset", "slow",
			"-g", "48",
			"-sc_threshold", "0",
			"-keyint_min", "48",
		)
	}

	args = append(args,
		"-f", "hls",
		"-hls_time", strconv.Itoa(segmentDuration),
		"-hls_playlist_type", "vod",
		"-hls_flags", "independent_segments",
		"-master_pl_name", masterPlaylist,
	)

	for _, res := range resolutions {
		args = append(args,
			"-var_stream_map", "v:"+res.Name+",a:0",
			"-hls_segment_filename", filepath.Join("output", res.Name, "segment_%03d.ts"),
			filepath.Join("output", res.Name, "playlist.m3u8"),
		)
	}

	// Create the command
	cmd := exec.Command("ffmpeg", args...)

	// Print the command for debugging
	fmt.Println("Executing FFmpeg command:")
	fmt.Println(strings.Join(cmd.Args, " "))

	return cmd
}
