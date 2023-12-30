package main

import (
	"fmt"
	"github.com/malc0mn/amiigo/amiibo"
	"github.com/malc0mn/amiigo/nfcptl"
	"sync"
	"time"
)

// portal holds the parts that need to be controlled by the NFC portal.
type portal struct {
	client *nfcptl.Client
	evt    chan struct{}          // Event channel used to re-init the portal.
	log    chan<- []byte          // Logger channel.
	amb    chan<- amiibo.Amiidump // Channel to send last read amiibo to.
	tkn    bool                   // Boolean to indicate a token is placed on the portal.
	con    bool                   // Boolean to indicate the portal is connected.
	sync.Mutex
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
				p.connected(true)
				p.log <- encodeStringCell(fmt.Sprintf("Successfully connected to NFC portal %04x:%04x.", p.client.VendorId(), p.client.ProductId()))
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
	//p.client.SendCommand(nfcptl.Command{Command: nfcptl.GetDeviceName})
	//p.client.SendCommand(nfcptl.Command{Command: nfcptl.GetHardwareInfo})

	for {
		select {
		case e := <-p.client.Events():
			p.log <- encodeStringCell(fmt.Sprintf("Received event: %s", e.String()))
			switch e.Name() {
			case nfcptl.TokenTagData:
				p.tokenState(true)
				a, err := amiibo.NewAmiibo(e.Data(), nil)
				if err != nil {
					p.log <- encodeStringCell(err.Error())
					continue
				}

				// Send amiibo to receiver
				p.amb <- a

				p.log <- encodeStringCell("NFC portal ready")
			case nfcptl.TokenRemoved:
				p.tokenState(false)
			case nfcptl.Disconnect:
				p.connected(false)
				p.log <- encodeStringCell("NFC portal disconnected!")
				p.reInit()
				return
			}
		case <-conf.quit:
			p.client.Disconnect()
			p.connected(false)
			return
		}
	}
}

// write sends amiibo data to the NFC portal which will then write it to the token placed on the
// portal. If user is true, then only the user data of the amiibo tag will be written,
// corresponding with 'restoring a backup' in the original software. Pass false to write the full
// amiibo data.
func (p *portal) write(data []byte, user bool) {
	if !p.isConnected() {
		p.log <- encodeStringCell("Cannot write: connect an NFC portal first!")
		return
	}

	if !p.tokenPlaced() {
		p.log <- encodeStringCell("Cannot write: please place a token on the NFC portal first!")
		return
	}

	typ := []byte{0x00}
	if user {
		typ = []byte{0x01}
	}

	p.log <- encodeStringCell("Sending amiibo data to NFC portal")
	p.client.SendCommand(nfcptl.Command{Command: nfcptl.WriteTokenData, Arguments: append(typ, data...)})
}

// reInit signals the receiver that the portal needs to be re-initialized due to a disconnect. The
// signal is sent from the listen function which should be running as a go routine. When this
// signal is sent, the listen go routine is stopped so that the receiver can simply start a new
// listen go routine to re-initialize the portal.
func (p *portal) reInit() {
	p.evt <- struct{}{}
}

// tokenState updates the token status in a thread safe way. True means a token is present on the
// portal.
func (p *portal) tokenState(tkn bool) {
	p.Lock()
	p.tkn = tkn
	p.Unlock()
}

// tokenPlaced returns true when a token is present on the NFC portal.
func (p *portal) tokenPlaced() bool {
	p.Lock()
	tkn := p.tkn
	p.Unlock()

	return tkn
}

// tokenState updates the token status in a thread safe way. True means a token is present on the
// portal.
func (p *portal) connected(con bool) {
	p.Lock()
	p.con = con
	p.Unlock()
}

// tokenPlaced returns true when a token is present on the NFC portal.
func (p *portal) isConnected() bool {
	p.Lock()
	con := p.con
	p.Unlock()

	return con
}

// newPortal returns a new portal ready for use.
func newPortal(log chan<- []byte, amiiboChan chan<- amiibo.Amiidump) *portal {
	return &portal{
		evt: make(chan struct{}),
		log: log,
		amb: amiiboChan,
	}
}
