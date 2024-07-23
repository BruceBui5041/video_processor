package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetVideoNames(folderPath string) ([]string, error) {
	// This command finds files with common video extensions
	cmd := exec.Command("find", folderPath, "-type", "f", "-name", "*.mp4", "-o", "-name", "*.avi", "-o", "-name", "*.mkv", "-o", "-name", "*.mov")

	output, err := cmd.Output()
	if err != nil {
		// Log the full error, including any output from the command
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Printf("Command failed with exit code %d. Stderr: %s", exitErr.ExitCode(), string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("error executing command: %v", err)
	}

	// Split the output into lines and extract file names
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
	// Get the base name of the file (in case it includes a path)
	base := filepath.Base(filename)

	// Find the last occurrence of the dot
	dotIndex := strings.LastIndex(base, ".")

	// If there's no dot, or it's the first character (hidden files in Unix),
	// return the original filename
	if dotIndex <= 0 {
		return base
	}

	// Return the part of the string before the last dot
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
		return nil, fmt.Errorf("error walking through directory: %v", err)
	}

	return filePaths, nil
}

func CreateDirIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
		fmt.Printf("Directory created successfully: %s\n", path)
	} else if err != nil {
		return fmt.Errorf("error checking directory: %v", err)
	} else {
		fmt.Printf("Directory already exists: %s\n", path)
	}
	return nil
}
