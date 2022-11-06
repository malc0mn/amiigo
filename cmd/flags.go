package main

import (
	"flag"
	"fmt"
)

var (
	showHelp    bool
	showVersion bool
	verbose     bool
	cFile       string
)

func initFlags() {
	flag.StringVar(&conf.vendor, "v", defaultVendor, "The vendor of the portal that will be connected to.")
	flag.StringVar(&conf.device, "d", defaultDevice, "The NFC portal to connect to.")
	flag.StringVar(&cFile, "c", "", "Read all settings from a config file. The config file will override any command line flags present.")

	flag.BoolVar(&verbose, "verbose", false, "Output lots and lots of debug information.")

	flag.BoolVar(&showHelp, "?", false, "Display usage information.")
	flag.BoolVar(&showVersion, "version", false, "Display version info.")

	flag.Usage = printUsage

	flag.Parse()
}

func printUsage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", exe)
	flag.PrintDefaults()
}
