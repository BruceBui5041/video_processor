package hlssegmenter

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Resolution struct {
	Width  int
	Height int
	Name   string
}

var resolutions = []Resolution{
	{Width: 1920, Height: 1080, Name: "1080p"},
	{Width: 1280, Height: 720, Name: "720p"},
	{Width: 854, Height: 480, Name: "480p"},
	{Width: 640, Height: 360, Name: "360p"},
}

const maxConcurrentProcesses = 4

func ExecHLSSegmentVideo(inputFile, outputDir string) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		log.Fatal("FFmpeg not found. Please install FFmpeg to continue.")
	}

	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	duration, err := getVideoDuration(inputFile)
	if err != nil {
		log.Fatalf("Failed to get video duration: %v", err)
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrentProcesses)
	var variantPlaylists []string
	var mu sync.Mutex

	for _, res := range resolutions {
		wg.Add(1)
		go func(res Resolution) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			resolutionDir := filepath.Join(outputDir, res.Name)
			if err := os.MkdirAll(resolutionDir, os.ModePerm); err != nil {
				log.Printf("Failed to create resolution directory for %s: %v", res.Name, err)
				return
			}

			playlistName := fmt.Sprintf("playlist_%s.m3u8", res.Name)
			cmd := generateFFmpegCommand(inputFile, resolutionDir, playlistName, 3, res)

			stderr, err := cmd.StderrPipe()
			if err != nil {
				log.Printf("Failed to create stderr pipe for %s: %v", res.Name, err)
				return
			}

			if err := cmd.Start(); err != nil {
				log.Printf("Failed to start FFmpeg for %s: %v", res.Name, err)
				return
			}

			go monitorProgress(stderr, duration, res.Name)

			if err := cmd.Wait(); err != nil {
				log.Printf("FFmpeg command failed for %s: %v", res.Name, err)
				return
			}

			mu.Lock()
			variantPlaylists = append(variantPlaylists, playlistName)
			mu.Unlock()

			fmt.Printf("\nHLS segmentation completed for %s.\n", res.Name)
		}(res)
	}

	wg.Wait()

	generateMasterPlaylist(outputDir, variantPlaylists)

	fmt.Printf("\nHLS segmentation completed successfully for all resolutions.\n")
	fmt.Printf("Output files are in the '%s' directory.\n", outputDir)
}

func generateMasterPlaylist(outputDir string, variantPlaylists []string) {
	masterPlaylistPath := filepath.Join(outputDir, "master.m3u8")
	f, err := os.Create(masterPlaylistPath)
	if err != nil {
		log.Fatalf("Failed to create master playlist: %v", err)
	}
	defer f.Close()

	f.WriteString("#EXTM3U\n")
	f.WriteString("#EXT-X-VERSION:3\n")

	for i, playlist := range variantPlaylists {
		res := resolutions[i]
		f.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d\n", getBandwidth(res), res.Width, res.Height))
		f.WriteString(fmt.Sprintf("%s/%s\n", res.Name, playlist))
	}
}

func getBandwidth(res Resolution) int {
	switch res.Name {
	case "1080p":
		return 5000000
	case "720p":
		return 2800000
	case "480p":
		return 1400000
	case "360p":
		return 800000
	default:
		return 400000
	}
}

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

func generateFFmpegCommand(inputFile, outputDir, playlistName string, segmentDuration int, res Resolution) *exec.Cmd {
	outputPath := filepath.Join(outputDir, "segment_%03d.ts")
	playlistPath := filepath.Join(outputDir, playlistName)

	args := []string{
		"-i", inputFile,
		"-profile:v", "main",
		"-level", "3.1",
		"-start_number", "0",
		"-hls_time", fmt.Sprintf("%d", segmentDuration),
		"-hls_list_size", "0",
		"-f", "hls",
		"-vf", fmt.Sprintf("scale=%d:%d", res.Width, res.Height),
		"-c:a", "aac",
		"-ar", "48000",
		"-b:a", "128k",
		"-hls_segment_filename", outputPath,
		playlistPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	fmt.Println("Executing FFmpeg command:")
	fmt.Println(strings.Join(cmd.Args, " "))
	return cmd
}

func monitorProgress(stderr io.ReadCloser, duration time.Duration, resName string) {
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
			fmt.Printf("\rProgress (%s): %.2f%%", resName, progress)
		}
	}
}
