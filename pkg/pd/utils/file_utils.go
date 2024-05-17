package utils

import (
	"log"
	"net/http"
	"os"
)

// GetFileSize returns the size of the file.
func GetFileSize(filePath string) int64 {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	return fileInfo.Size()
}

// GetMimeType returns the MIME type of the file.
func GetMimeType(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return ""
	}

	return http.DetectContentType(buffer)
}
