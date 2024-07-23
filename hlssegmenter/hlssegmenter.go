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
	"video_processor/constants"
	"video_processor/pubsub"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
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
const videoSegmentDuration = 3

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
	variantPlaylists := make([]string, len(resolutions))
	var mu sync.Mutex

	for i, res := range resolutions {
		wg.Add(1)
		go func(i int, res Resolution) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			resolutionDir := filepath.Join(outputDir, res.Name)
			if err := os.MkdirAll(resolutionDir, os.ModePerm); err != nil {
				log.Printf("Failed to create resolution directory for %s: %v", res.Name, err)
				return
			}

			playlistName := fmt.Sprintf("playlist_%s.m3u8", res.Name)
			cmd, err := generateFFmpegCommand(inputFile, resolutionDir, playlistName, videoSegmentDuration, res)
			if err != nil {
				log.Printf("Failed to generate FFmpeg command for %s: %v", res.Name, err)
				return
			}

			stderrPipe, err := cmd.StderrPipe()
			if err != nil {
				log.Printf("Failed to create stderr pipe for %s: %v", res.Name, err)
				return
			}

			fmt.Printf("Starting FFmpeg for %s\n", res.Name)

			if err := cmd.Start(); err != nil {
				log.Printf("Failed to start FFmpeg for %s: %v", res.Name, err)
				return
			}

			go monitorProgress(stderrPipe, duration, res.Name)

			if err := cmd.Wait(); err != nil {
				log.Printf("FFmpeg command failed for %s: %v", res.Name, err)
				return
			}

			fmt.Printf("\nFFmpeg completed successfully for %s\n", res.Name)

			mu.Lock()
			variantPlaylists[i] = playlistName
			fmt.Printf("Added playlist for %s at index %d: %s\n", res.Name, i, playlistName)
			mu.Unlock()

			fmt.Printf("HLS segmentation completed for %s.\n", res.Name)
		}(i, res)
	}

	wg.Wait()

	fmt.Println("Final variant playlists:")
	for i, playlist := range variantPlaylists {
		fmt.Printf("Index %d: %s\n", i, playlist)
	}

	generateMasterPlaylist(outputDir, variantPlaylists)

	fmt.Printf("\nHLS segmentation completed successfully for all resolutions.\n")
	fmt.Printf("Output files are in the '%s' directory.\n", outputDir)

	// Publish event when processing is complete
	msg := message.NewMessage(watermill.NewUUID(), []byte(fmt.Sprintf("%s,%s", inputFile, outputDir)))
	if err := pubsub.Publisher.Publish(constants.TopicVideoProcessed, msg); err != nil {
		log.Printf("Failed to publish video_processed event: %v", err)
	}
}

func generateMasterPlaylist(outputDir string, variantPlaylists []string) {
	fmt.Println("Generating master playlist with:", variantPlaylists)

	masterPlaylistPath := filepath.Join(outputDir, "master.m3u8")
	f, err := os.Create(masterPlaylistPath)
	if err != nil {
		log.Fatalf("Failed to create master playlist: %v", err)
	}
	defer f.Close()

	f.WriteString("#EXTM3U\n")
	f.WriteString("#EXT-X-VERSION:3\n")

	for i, playlist := range variantPlaylists {
		if playlist == "" {
			fmt.Printf("Empty playlist at index %d\n", i)
			continue
		}
		res := resolutions[i]
		entry := fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d\n%s/%s\n",
			getBandwidth(res), res.Width, res.Height, res.Name, playlist)
		f.WriteString(entry)
		fmt.Printf("Added to master playlist: %s", entry)
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

func generateFFmpegCommand(inputFile, outputDir, playlistName string, segmentDuration int, res Resolution) (*exec.Cmd, error) {
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
		"-force_key_frames", fmt.Sprintf("expr:gte(t,n_forced*%d)", segmentDuration),
		"-hls_flags", "split_by_time+independent_segments",
		"-hls_segment_type", "mpegts",
		"-hls_segment_filename", outputPath,
		playlistPath,
	}

	cmd := exec.Command("ffmpeg", args...)

	return cmd, nil
}

func monitorProgress(stderr io.Reader, duration time.Duration, resName string) {
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
			// Flush the output to ensure it's displayed immediately
			fmt.Print("\033[?25l") // Hide cursor
		}
	}
	fmt.Print("\033[?25h") // Show cursor
}
