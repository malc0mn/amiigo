package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	ok               = 0
	errGeneral       = 1
	errInvalidArgs   = 2
	errOpenConfig    = 102
	errCreateClient  = 104
	errPortalConnect = 105
)

var (
	version   = "0.0.0"
	buildTime = "unknown"
	exe       string
	quit      = make(chan struct{})
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

	log.SetOutput(logging{log.Writer()})

	if err := createCacheDirs(); err != nil {
		fmt.Printf("Error creating caching directories: %s\n", err)
		os.Exit(errGeneral)
	}
	tui()

	fmt.Println("Bye bye!")
	os.Exit(ok)
}
