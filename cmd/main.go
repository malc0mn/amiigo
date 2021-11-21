package main

import (
	"fmt"
	"github.com/malc0mn/amiigo/usb"
	"log"
	"os"
	"path/filepath"
)

const (
	ok = 0
)

var (
	version   = "0.0.0"
	buildTime = "unknown"
	exe       string
)

func main() {
	exe = filepath.Base(os.Args[0])

	initFlags()
	if showHelp {
		printUsage()
		os.Exit(ok)
	}

	if showVersion {
		fmt.Printf("%s version %s built on %s\n", exe, version, buildTime)
		os.Exit(ok)
	}

	client, err := usb.NewClient(conf.vendor, conf.device, verbose)
	if err != nil {
		log.Fatalf("Error initialising client: %s", err)
	}

	err = client.Connect()
	defer client.Disconnect()
	if err != nil {
		log.Fatalf("Error connecting to device: %v", err)
	}
}
