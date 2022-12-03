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
	errOpenConfig    = 102
	errOpenLogFile   = 103
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

	if cFile != "" {
		if err := loadConfig(cFile, conf); err != nil {
			fmt.Fprintf(os.Stderr, "Error opening config file - %s\n", err)
			os.Exit(errOpenConfig)
		}
	}

	f, err := getLogfile(conf.logFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening log file - %s\n", err)
		os.Exit(errOpenLogFile)
	}
	defer f.Close()
	log.SetOutput(f)

	if err := createCacheDirs(conf.cacheDir); err != nil {
		fmt.Printf("Error creating caching directories: %s\n", err)
		os.Exit(errGeneral)
	}

	tui(conf)
	conf.wg.Wait()

	fmt.Println("Bye bye!")
	os.Exit(ok)
}
