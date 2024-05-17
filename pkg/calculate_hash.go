package main

import (
	"fmt"
	"github.com/itsDarianNgo/go-pd/pkg/pd/utils"
	"log"
)

func main() {
	files := []string{
		"pd/testdata/cat.jpg",
		"pd/testdata/cat_unique1.jpg",
		"pd/testdata/cat_unique2.jpg",
		"pd/testdata/cat_unique3.jpg",
		"pd/testdata/cat_unique4.jpg",
		"pd/testdata/cat_unique5.jpg",
		"pd/testdata/cat_unique6.jpg",
	}

	for _, filePath := range files {
		hash, err := utils.CalculateFileHash(filePath)
		if err != nil {
			log.Fatalf("Failed to calculate hash for %s: %v", filePath, err)
		}
		fmt.Printf("SHA-256 Hash of %s: %s\n", filePath, hash)
	}
}
