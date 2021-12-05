package main

import (
	"encoding/hex"
	"fmt"
	"github.com/malc0mn/amiigo/nfcptl"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

const (
	ok = 0
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

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Printf("Received signal %s, shutting down...\n", sig)
		close(quit)
	}()

	client, err := nfcptl.NewClient(conf.vendor, conf.device, quit, verbose)
	if err != nil {
		log.Fatalf("Error initialising client: %s", err)
	}

	err = client.Connect()
	defer client.Disconnect()
	if err != nil {
		log.Fatalf("Error connecting to device: %v", err)
	}

	//client.SendCommand(nfcptl.GetDeviceName)
	//client.SendCommand(nfcptl.GetHardwareInfo)

	for {
		select {
		case e := <-client.Events():
			fmt.Println("Received event:")
			fmt.Println(e.String())
			fmt.Println(hex.Dump(e.Data()))
		case <-quit:
			return
		}
	}

	fmt.Println("Bye bye!")
}
