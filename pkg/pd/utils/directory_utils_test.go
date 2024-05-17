package utils

import (
	"testing"
)

func TestGetFilesInDirectory(t *testing.T) {
	files, err := GetFilesInDirectory("../testdata/test_directory")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(files) == 0 {
		t.Fatalf("Expected files, got none")
	}

	// Check that the correct number of files are found
	expectedFilesCount := 3
	if len(files) != expectedFilesCount {
		t.Fatalf("Expected %d files, got %d", expectedFilesCount, len(files))
	}
}
