package main

import (
	"encoding/hex"
	"fmt"
	"github.com/malc0mn/amiigo/amiibo"
	"github.com/malc0mn/amiigo/apii"
	"github.com/malc0mn/amiigo/nfcptl"
	"time"
)

// portal holds the parts that need to be controlled by the NFC portal.
type portal struct {
	client *nfcptl.Client
	evt    chan struct{}   // Event channel used to re-init the portal.
	log    chan<- []byte   // Logger channel.
	ifo    chan<- []byte   // Info channel.
	usg    chan<- []byte   // Usage channel.
	img    *imageBox       // The image box.
	api    *apii.AmiiboAPI // An amiibo HTTP API instance.
}

// connect will block until a successful connection is established to the USB NFC portal.
func (p *portal) connect(quit <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	output := false

	for {
		select {
		case <-quit:
			return

		case <-ticker.C:
			err := p.client.Connect()
			if err == nil {
				p.log <- encodeStringCell("Successfully connected to NFC portal.")
				return
			}
			if !output {
				p.log <- encodeStringCell("Please connect your amiibo NFC portal to a USB port.")
				output = true
			}
		}
	}
}

// listen is the main hardware loop that handles all portal related things from connection over
// handling its events to cleanly disconnecting on shutdown.
func (p *portal) listen(conf *config) {
	conf.wg.Add(1)
	defer conf.wg.Done()

	var err error
	p.client, err = nfcptl.NewClient(conf.vendor, conf.device, verbose)
	if err != nil {
		p.log <- encodeStringCell(fmt.Sprintf("Error initialising client: %s\n", err))
		return
	}

	p.connect(conf.quit)

	// TODO: Send this output to the UI, maybe close to the logo somewhere?
	//client.SendCommand(nfcptl.GetDeviceName)
	//client.SendCommand(nfcptl.GetHardwareInfo)

	for {
		select {
		case e := <-p.client.Events():
			p.log <- encodeStringCell(fmt.Sprintf("Received event: %s", e.String()))
			if e.Name() == nfcptl.TokenTagData {
				a, err := amiibo.NewAmiibo(e.Data(), nil)
				if err != nil {
					p.log <- encodeStringCell(err.Error())
					continue
				}

				rawId := a.ModelInfo().ID()
				if zeroed(rawId) {
					continue
				}
				id := hex.EncodeToString(rawId)
				p.log <- encodeStringCell("Got id: " + id)

				// Fill info box.
				p.log <- encodeStringCell("Fetching amiibo info")
				ai, err := p.api.GetAmiiboInfoById(id)
				if err != nil {
					p.log <- encodeStringCell("API get amiibo info: " + err.Error())
					continue
				}
				p.ifo <- formatAmiiboInfo(ai)

				// Fill image box.
				p.log <- encodeStringCell("Fetching image")
				img, err := getImage(ai.Image)
				if err != nil {
					p.log <- encodeStringCell("API get image: " + err.Error())
					continue
				}
				p.img.setImage(img)

				// Fill usage box.
				p.log <- encodeStringCell("Fetching character usage")
				cu, err := p.api.GetCharacterUsage(ai.Character)
				if err != nil {
					p.log <- encodeStringCell("Api get character usage: " + err.Error())
					continue
				}
				p.usg <- formatAmiiboUsage(cu)

				p.log <- encodeStringCell("Ready")
			} else if e.Name() == nfcptl.Disconnect {
				p.log <- encodeStringCell("NFC portal disconnected!")
				p.reInit()
				return
			}
		case <-conf.quit:
			p.client.Disconnect()
			return
		}
	}
}

// reInit signals the receiver that the portal needs to be re-initialized due to a disconnect. The
// signal is sent from the listen function which should be running as a go routine. When this
// signal is sent, the listen go routine is stopped so that the receiver can simply start a new
// listen go routine to re-initialize the portal.
func (p *portal) reInit() {
	p.evt <- struct{}{}
}

// newPortal returns a new portal ready for use.
func newPortal(log, ifo, usg chan<- []byte, img *imageBox, baseUrl string) *portal {
	return &portal{
		evt: make(chan struct{}),
		log: log,
		ifo: ifo,
		usg: usg,
		img: img,
		api: apii.NewAmiiboAPI(newCachedHttpClient(), baseUrl),
	}
}
