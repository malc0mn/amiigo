package main

import (
	"encoding/hex"
	"fmt"
	"github.com/malc0mn/amiigo/amiibo"
	"github.com/malc0mn/amiigo/apii"
	"github.com/malc0mn/amiigo/nfcptl"
)

// portal holds the parts that need to be controlled by the NFC portal.
type portal struct {
	log chan<- []byte
	img *imageBox
	api *apii.AmiiboAPI
}

// listen is the main hardware loop that handles all portal related things from connection over
// handling its events to cleanly disconnecting on shutdown.
func (p *portal) listen(conf *config) {
	client, err := nfcptl.NewClient(conf.vendor, conf.device, verbose)
	if err != nil {
		p.log <- encodeStringCell(fmt.Sprintf("Error initialising client: %s\n", err))
		return
	}

	err = client.Connect()
	defer client.Disconnect()
	if err != nil {
		p.log <- encodeStringCell(fmt.Sprintf("Error connecting to device: %s\n", err))
		return
	}

	// TODO: Send this output to the UI, maybe close to the logo somewhere?
	//client.SendCommand(nfcptl.GetDeviceName)
	//client.SendCommand(nfcptl.GetHardwareInfo)

	for {
		select {
		case e := <-client.Events():
			p.log <- encodeStringCell(fmt.Sprintf("Received event: %s", e.String()))
			if e.Name() == nfcptl.TokenTagData {
				a, err := amiibo.NewAmiibo(e.Data(), nil)
				if err != nil {
					p.log <- encodeStringCell(err.Error())
					continue
				}

				id := hex.EncodeToString(a.ModelInfo().ID())
				p.log <- encodeStringCell("got id: " + id)

				ai, err := p.api.GetAmiiboInfoById(id)
				if err != nil {
					p.log <- encodeStringCell(err.Error())
					continue
				}

				img, err := getImage(ai.Image)
				if err != nil {
					p.log <- encodeStringCell(err.Error())
					continue
				}

				p.img.processImage(img)
			}
		case <-quit:
			// TODO: we must somehow wait for a clean driver shutdown before quitting.
			return
		}
	}
}

// newPortal returns a new portal ready for use.
func newPortal(log chan<- []byte, img *imageBox, baseUrl string) *portal {
	return &portal{
		log: log,
		img: img,
		api: apii.NewAmiiboAPI(newCachedHttpClient(), baseUrl),
	}
}
