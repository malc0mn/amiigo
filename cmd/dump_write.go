package main

import (
	"fmt"
	"github.com/malc0mn/amiigo/amiibo"
	"os"
	"path/filepath"
)

// writeDump writes the active amiibo data to disk.
func writeDump(filename string, a amiibo.Amiidump, log chan<- []byte) bool {
	if a == nil {
		log <- encodeStringCell("No amiibo data to write!")
		return false
	}
	if filename == "" {
		log <- encodeStringCell("Please provide a filename!")
		return false
	}

	if ext := filepath.Ext(filename); ext != ".bin" {
		filename += ".bin"
	}

	dest := filename
	dir := filepath.Dir(filename)
	if dir == "." {
		dir, _ = os.Getwd()
		dest = filepath.Join(dir, filename)
	}

	log <- encodeStringCell(fmt.Sprintf("Writing amiibo to file '%s'", dest))
	if err := os.WriteFile(filename, a.Raw(), 0644); err != nil {
		log <- encodeStringCell(fmt.Sprintf("Error writing file: %s", err))
		return false
	}

	log <- encodeStringCell("Amiibo dump successful!")
	return true
}
