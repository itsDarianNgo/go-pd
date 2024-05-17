package utils

import (
	"encoding/csv"
	"os"
)

// UploadInfo holds the information about the uploaded file.
type UploadInfo struct {
	FileName       string `csv:"file_name"`
	DirectoryPath  string `csv:"directory_path"`
	URL            string `csv:"url"`
	UploadDateTime string `csv:"upload_date_time"`
	FileSize       int64  `csv:"file_size"`
	FormattedSize  string `csv:"formatted_size"`
	MIMEType       string `csv:"mime_type"`
	Uploader       string `csv:"uploader"`
	UploadStatus   string `csv:"upload_status"`
}

// SaveUploadInfoToCSV saves the upload information to a CSV file.
func SaveUploadInfoToCSV(info UploadInfo, filePath string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	record := []string{
		info.FileName,
		info.DirectoryPath,
		info.URL,
		info.UploadDateTime,
		FormatFileSize(info.FileSize), // Use the formatted size here
		info.MIMEType,
		info.Uploader,
		info.UploadStatus,
	}

	return writer.Write(record)
}
