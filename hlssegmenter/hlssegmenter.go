package hlssegmenter

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"video_processor/appconst"
	"video_processor/logger"
	"video_processor/storagehandler"
	"video_processor/utils"

	"go.uber.org/zap"
)

type Resolution struct {
	Width           int
	Height          int
	Name            string
	SegmentDuration int // New field for segment duration
}

var resolutions = []Resolution{
	{Width: 1920, Height: 1080, Name: "1080p", SegmentDuration: 2},
	{Width: 1280, Height: 720, Name: "720p", SegmentDuration: 3},
	{Width: 854, Height: 480, Name: "480p", SegmentDuration: 4},
	{Width: 640, Height: 360, Name: "360p", SegmentDuration: 5},
}

const maxConcurrentProcesses = 4

func StartSegmentProcess(inputFile, outputDir string) (string, string, error) {
	utils.CreateDirIfNotExist(appconst.UnprecessedVideoDir)
	unprecessedVideoPath, err := storagehandler.GetS3File(
		appconst.AWSVideoS3BuckerName,
		inputFile,
		appconst.UnprecessedVideoDir,
	)
	if err != nil {
		logger.AppLogger.Error("Failed to get S3 file", zap.Error(err), zap.String("inputFile", inputFile))
		return "", "", err
	}

	desireOutputPath := filepath.Join(outputDir, filepath.Base(inputFile))
	utils.CreateDirIfNotExist(desireOutputPath)
	return hslSegmentVideo(unprecessedVideoPath, desireOutputPath)
}

func hslSegmentVideo(inputFile, outputDir string) (string, string, error) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		logger.AppLogger.Error("FFmpeg not found. Please install FFmpeg to continue.", zap.Error(err))
		return "", "", err
	}

	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		logger.AppLogger.Error("Failed to create output directory", zap.Error(err), zap.String("outputDir", outputDir))
		return "", "", err
	}

	// Extract video name from inputFile
	videoName := filepath.Base(inputFile)

	duration, err := getVideoDuration(inputFile)
	if err != nil {
		logger.AppLogger.Error("Failed to get video duration", zap.Error(err), zap.String("inputFile", inputFile))
		return "", "", err
	}

	variantPlaylists := make([]string, len(resolutions))
	var mu sync.Mutex

	for i, res := range resolutions {

		resolutionDir := filepath.Join(outputDir, res.Name)
		if err := os.MkdirAll(resolutionDir, os.ModePerm); err != nil {
			logger.AppLogger.Error("Failed to create resolution directory",
				zap.Error(err),
				zap.String("resolution", res.Name),
				zap.String("dir", resolutionDir))
			return "", "", err
		}

		playlistName := fmt.Sprintf("playlist_%s.m3u8", res.Name)
		cmd, err := generateFFmpegCommand(inputFile, resolutionDir, playlistName, res)
		if err != nil {
			logger.AppLogger.Error("Failed to generate FFmpeg command",
				zap.Error(err),
				zap.String("resolution", res.Name))
			return "", "", err
		}

		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			logger.AppLogger.Error("Failed to create stderr pipe",
				zap.Error(err),
				zap.String("resolution", res.Name))
			return "", "", err
		}

		logger.AppLogger.Info("Starting FFmpeg", zap.String("resolution", res.Name))

		if err := cmd.Start(); err != nil {
			logger.AppLogger.Error("Failed to start FFmpeg",
				zap.Error(err),
				zap.String("resolution", res.Name))
			return "", "", err
		}

		go monitorProgress(stderrPipe, duration, res.Name)

		if err := cmd.Wait(); err != nil {
			logger.AppLogger.Error("FFmpeg command failed",
				zap.Error(err),
				zap.String("resolution", res.Name))
			return "", "", err
		}

		logger.AppLogger.Info("FFmpeg completed successfully", zap.String("resolution", res.Name))

		mu.Lock()
		variantPlaylists[i] = playlistName
		logger.AppLogger.Info("Added playlist",
			zap.String("resolution", res.Name),
			zap.Int("index", i),
			zap.String("playlist", playlistName))
		mu.Unlock()

		logger.AppLogger.Info("HLS segmentation completed", zap.String("resolution", res.Name))
	}

	logger.AppLogger.Info("Final variant playlists", zap.Strings("playlists", variantPlaylists))

	generateMasterPlaylist(outputDir, variantPlaylists, videoName)

	logger.AppLogger.Info("HLS segmentation completed successfully for all resolutions",
		zap.String("outputDir", outputDir))

	return inputFile, outputDir, nil
}

func generateMasterPlaylist(outputDir string, variantPlaylists []string, videoName string) {
	logger.AppLogger.Info("Generating master playlist", zap.Strings("variantPlaylists", variantPlaylists))

	masterPlaylistPath := filepath.Join(outputDir, "master.m3u8")
	f, err := os.Create(masterPlaylistPath)
	if err != nil {
		logger.AppLogger.Fatal("Failed to create master playlist",
			zap.Error(err),
			zap.String("path", masterPlaylistPath))
	}
	defer f.Close()

	f.WriteString("#EXTM3U\n")
	f.WriteString("#EXT-X-VERSION:3\n")

	for i, playlist := range variantPlaylists {
		if playlist == "" {
			logger.AppLogger.Warn("Empty playlist", zap.Int("index", i))
			continue
		}
		res := resolutions[i]
		entry := fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d\n%s/%s/%s\n",
			getBandwidth(res), res.Width, res.Height, videoName, res.Name, playlist)
		f.WriteString(entry)
		logger.AppLogger.Info("Added to master playlist",
			zap.String("entry", entry),
			zap.String("resolution", res.Name))
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

func generateFFmpegCommand(inputFile, outputDir, playlistName string, res Resolution) (*exec.Cmd, error) {
	outputPath := filepath.Join(outputDir, "segment_%03d.ts")
	playlistPath := filepath.Join(outputDir, playlistName)

	args := []string{
		"-i", inputFile,
		"-profile:v", "main",
		"-level", "3.1",
		"-start_number", "0",
		"-hls_time", fmt.Sprintf("%d", res.SegmentDuration),
		"-hls_list_size", "0",
		"-f", "hls",
		"-vf", fmt.Sprintf("scale=%d:%d", res.Width, res.Height),
		"-c:a", "aac",
		"-ar", "48000",
		"-b:a", "128k",
		"-force_key_frames", fmt.Sprintf("expr:gte(t,n_forced*%d)", res.SegmentDuration),
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
			logger.AppLogger.Info("Progress",
				zap.String("resolution", resName),
				zap.Float64("percentage", progress))
		}
	}
}
