package utils

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// GetHashFilePath returns the appropriate hash file path based on the environment mode.
func GetHashFilePath() string {
	envMode := os.Getenv("ENV_MODE")
	if envMode == "test" {
		return "test_hashes.csv"
	}
	return "hashes.csv"
}

// CalculateFileHash calculates and returns the SHA-256 hash of a file.
func CalculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			fmt.Printf("Error closing file: %v\n", cerr)
		}
	}()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// InitializeHashFile checks if the hash file exists and creates it if not.
func InitializeHashFile(hashFilePath string) error {
	if _, err := os.Stat(hashFilePath); os.IsNotExist(err) {
		file, err := os.Create(hashFilePath)
		if err != nil {
			return err
		}
		if cerr := file.Close(); cerr != nil {
			return cerr
		}
	}
	return nil
}

// SaveFileHash saves the file path and its hash to a CSV file if it doesn't already exist.
func SaveFileHash(hashFilePath, filePath, hash string) error {
	if err := InitializeHashFile(hashFilePath); err != nil {
		return err
	}

	// Check if the file is a duplicate before saving
	isDuplicate, err := IsDuplicate(hashFilePath, filePath)
	if err != nil {
		return err
	}
	if isDuplicate {
		return nil // Do not save if the file is a duplicate
	}

	file, err := os.OpenFile(hashFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			fmt.Printf("Error closing file: %v\n", cerr)
		}
	}()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	return writer.Write([]string{filePath, hash})
}

// LoadFileHashes loads the file hashes from a CSV file into a map.
func LoadFileHashes(hashFilePath string) (map[string]string, error) {
	if err := InitializeHashFile(hashFilePath); err != nil {
		return nil, err
	}

	file, err := os.Open(hashFilePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			fmt.Printf("Error closing file: %v\n", cerr)
		}
	}()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	hashes := make(map[string]string)
	for _, record := range records {
		hashes[record[0]] = record[1]
	}

	return hashes, nil
}

// IsDuplicate checks if the file is a duplicate by comparing its hash with stored hashes.
func IsDuplicate(hashFilePath, filePath string) (bool, error) {
	newHash, err := CalculateFileHash(filePath)
	if err != nil {
		return false, err
	}

	hashes, err := LoadFileHashes(hashFilePath)
	if err != nil {
		return false, err
	}

	for _, hash := range hashes {
		if hash == newHash {
			return true, nil
		}
	}

	return false, nil
}

// PrintFileHash prints the SHA-256 hash of a given file.
func PrintFileHash(filePath string) {
	hash, err := CalculateFileHash(filePath)
	if err != nil {
		fmt.Printf("Failed to calculate hash: %v\n", err)
		return
	}

	fmt.Printf("SHA-256 Hash: %s\n", hash)
}
