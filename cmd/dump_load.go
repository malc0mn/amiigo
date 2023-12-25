package main

import (
	"fmt"
	"github.com/malc0mn/amiigo/amiibo"
	"os"
	"path/filepath"
)

// loadDump loads an amiibo dump from disk.
func loadDump(filename string, _ *amiibo.Amiibo, log chan<- []byte) bool {
	if filename == "" {
		log <- encodeStringCell("Please provide a filename!")
		return false
	}

	src := filename
	dir := filepath.Dir(filename)
	if dir == "." {
		dir, _ = os.Getwd()
		src = filepath.Join(dir, filename)
	}

	log <- encodeStringCell(fmt.Sprintf("Reading amiibo from file '%s'", src))

	data, err := os.ReadFile(filename)
	if err != nil {
		log <- encodeStringCell(fmt.Sprintf("Error reading file: %s", err))
		return false
	}

	am, err := amiibo.NewAmiibo(data, nil)
	if err != nil {
		log <- encodeStringCell(fmt.Sprintf("Error reading amiibo data: %s", err))
		return false
	}

	amiiboChan <- *am

	log <- encodeStringCell("Amiibo read successful!")
	return true
}
