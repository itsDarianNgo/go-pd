package utils

import (
	"log"
	"os"
	"path/filepath"
)

// GetFilesInDirectory recursively collects file paths in a directory
func GetFilesInDirectory(dirPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			log.Printf("Found file: %s", path)
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
