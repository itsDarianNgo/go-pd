package utils

import (
	"encoding/csv"
	"fmt"
	"os"
)

// UploadInfo holds the information about the uploaded file.
type UploadInfo struct {
	FileName       string
	DirectoryPath  string
	URL            string
	UploadDateTime string
	FileSize       int64
	MIMEType       string
	Uploader       string
	UploadStatus   string
}

// SaveUploadInfoToCSV saves the upload information to a CSV file.
func SaveUploadInfoToCSV(info UploadInfo, csvPath string) error {
	file, err := os.OpenFile(csvPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header if the file is new
	if fi, err := file.Stat(); err == nil && fi.Size() == 0 {
		if err := writer.Write([]string{"File Name", "Directory Path", "URL", "Upload Date and Time", "File Size", "MIME Type", "Uploader Username", "Upload Status"}); err != nil {
			return err
		}
	}

	record := []string{
		info.FileName,
		info.DirectoryPath,
		info.URL,
		info.UploadDateTime,
		fmt.Sprintf("%d", info.FileSize),
		info.MIMEType,
		info.Uploader,
		info.UploadStatus,
	}

	if err := writer.Write(record); err != nil {
		return err
	}

	return nil
}
