package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"video_processor/logger"

	"go.uber.org/zap"
)

func GetVideoNames(folderPath string) ([]string, error) {
	cmd := exec.Command("find", folderPath, "-type", "f", "-name", "*.mp4", "-o", "-name", "*.avi", "-o", "-name", "*.mkv", "-o", "-name", "*.mov")

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			logger.AppLogger.Error("Command failed",
				zap.Int("exitCode", exitErr.ExitCode()),
				zap.String("stderr", string(exitErr.Stderr)))
		}
		return nil, fmt.Errorf("error executing command: %v", err)
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	var videoNames []string
	for _, file := range files {
		parts := strings.Split(file, "/")
		if len(parts) > 0 {
			videoNames = append(videoNames, parts[len(parts)-1])
		}
	}

	return videoNames, nil
}

func RemoveFileExtension(filename string) string {
	base := filepath.Base(filename)
	dotIndex := strings.LastIndex(base, ".")
	if dotIndex <= 0 {
		return base
	}
	return base[:dotIndex]
}

func GetFilePaths(dirPath string) ([]string, error) {
	var filePaths []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			filePaths = append(filePaths, path)
		}
		return nil
	})

	if err != nil {
		logger.AppLogger.Error("Error walking through directory", zap.Error(err), zap.String("directory", dirPath))
		return nil, fmt.Errorf("error walking through directory: %v", err)
	}

	return filePaths, nil
}

func CreateDirIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			logger.AppLogger.Error("Failed to create directory", zap.Error(err), zap.String("path", path))
			return fmt.Errorf("failed to create directory: %v", err)
		}
		logger.AppLogger.Info("Directory created successfully", zap.String("path", path))
	} else if err != nil {
		logger.AppLogger.Error("Error checking directory", zap.Error(err), zap.String("path", path))
		return fmt.Errorf("error checking directory: %v", err)
	} else {
		logger.AppLogger.Info("Directory already exists", zap.String("path", path))
	}
	return nil
}

func DeleteLocalFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		logger.AppLogger.Error("Failed to delete file", zap.Error(err), zap.String("path", path))
		return fmt.Errorf("failed to delete file %s: %v", path, err)
	}
	logger.AppLogger.Info("Successfully deleted file", zap.String("path", path))
	return nil
}

func DeleteDirContents(dirPath string) error {
	dir, err := os.Open(dirPath)
	if err != nil {
		logger.AppLogger.Error("Failed to open directory", zap.Error(err), zap.String("path", dirPath))
		return fmt.Errorf("failed to open directory %s: %v", dirPath, err)
	}
	defer dir.Close()

	entries, err := dir.Readdirnames(-1)
	if err != nil {
		logger.AppLogger.Error("Failed to read directory contents", zap.Error(err), zap.String("path", dirPath))
		return fmt.Errorf("failed to read directory contents of %s: %v", dirPath, err)
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry)
		err = os.RemoveAll(fullPath)
		if err != nil {
			logger.AppLogger.Error("Failed to remove item", zap.Error(err), zap.String("path", fullPath))
			return fmt.Errorf("failed to remove %s: %v", fullPath, err)
		}
	}

	logger.AppLogger.Info("Successfully deleted all contents of directory", zap.String("path", dirPath))
	return nil
}
