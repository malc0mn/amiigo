package main

import (
	"encoding/hex"
	"fmt"
	"github.com/malc0mn/amiigo/nfcptl"
	"os"
)

// portal is the main hardware loop that handles all portal related things from connection over
// handling its events to cleanly disconnecting on shutdown.
func portal(log chan<- string) {
	client, err := nfcptl.NewClient(conf.vendor, conf.device, verbose)
	if err != nil {
		log <- fmt.Sprintf("Error initialising client: %s\n", err)
		return
	}

	err = client.Connect()
	defer client.Disconnect()
	if err != nil {
		log <- fmt.Sprintf("Error connecting to device: %v\n", err)
		return
	}

	//client.SendCommand(nfcptl.GetDeviceName)
	//client.SendCommand(nfcptl.GetHardwareInfo)

	for {
		select {
		case e := <-client.Events():
			log <- fmt.Sprintf("Received event: %s", e.String())
			fmt.Fprintln(os.Stderr, hex.Dump(e.Data()))
			if e.Name() == nfcptl.TokenTagData {
				os.WriteFile("ps_dump.bin", e.Data(), 0666)
			}
		case <-quit:
			return
		}
	}
}
