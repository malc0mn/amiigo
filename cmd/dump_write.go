package main

import (
	"fmt"
	"github.com/malc0mn/amiigo/amiibo"
	"os"
	"path/filepath"
)

// writeDump writes the active amiibo data to disk.
func writeDump(filename string, a *amiibo.Amiibo, log chan<- []byte) {
	if a == nil {
		log <- encodeStringCell("No amiibo data to write!")
		return
	}
	if filename == "" {
		log <- encodeStringCell("Please provide a filename!")
		return
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
		return
	}

	log <- encodeStringCell("Amiibo dump successful!")
}
