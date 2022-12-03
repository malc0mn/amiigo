package nfcptl

import (
	"log"
	"sync"
)

// Client allows easy communications with an NFC portal connected over USB.
type Client struct {
	va string // The vendor alias to use
	pa string // The product alias to use

	driver Driver // The driver to use for communicating with the NFC portal

	wg sync.WaitGroup // The client will use this to wait for the driver's goroutine(s) to finish.

	debug bool // Will enable verbose logging

	// TODO: do we need to use a context instead of a terminate channel?
	terminate chan struct{} // A channel telling the Driver to terminate as soon as the channel is closed.
	events    chan *Event   // Device events will be received on this channel.
	commands  chan Command  // Command structs will be sent on this channel for the driver to act upon.
}

// NewClient builds a new Client struct.
func NewClient(vendor, device string, debug bool) (*Client, error) {
	// TODO: add auto detection when vendor and device are empty strings.
	d, err := GetDriverByVendorAndProductAlias(vendor, device)
	if err != nil {
		return nil, err
	}

	c := &Client{
		va:        vendor,
		pa:        device,
		driver:    d,
		debug:     debug,
		terminate: make(chan struct{}),
		events:    make(chan *Event, 10),
		commands:  make(chan Command, 1),
	}

	if c.Debug() {
		log.Printf("nfcptl: using vendor ID %#04x and product ID %#04x", c.VendorId(), c.ProductId())
	}

	return c, nil
}

// Setup returns the driver's setup struct. This will be protocol dependant: the protocol will
// verify the setup struct and panic when it is not what it expects.
func (c *Client) Setup() any {
	return c.driver.Setup()
}

// Connect establishes a new connection to the device and opens up input and output endpoints.
func (c *Client) Connect() error {
	if err := c.driver.Connect(c); err != nil {
		return err
	}

	// Pass control to the driver now.
	c.wg.Add(1)
	go c.driver.Drive(c)

	return nil
}

// Done MUST be used by drivers to signal the end of the main goroutine which will allow the client
// to ensure a clean shutdown.
func (c *Client) Done() {
	c.wg.Done()
}

// Disconnect cleanly disconnects the client and frees up all resources.
func (c *Client) Disconnect() error {
	close(c.terminate)

	// Wait for a clean shutdown of the driver's goroutines.
	c.wg.Wait()

	if err := c.driver.Disconnect(); err != nil {
		return err
	}

	close(c.events)

	return nil
}

// Events returns the read only events channel. The caller MUST use this channel to listen for
// device events and act accordingly.
func (c *Client) Events() <-chan *Event {
	return c.events
}

// PublishEvent places an event on the event channel.
// This function is exposed to allow Driver implementations outside the nfcptl package.
func (c *Client) PublishEvent(e *Event) {
	c.events <- e
}

// Terminate returns the termination channel which a Driver MUST use to cleanly terminate any
// goroutines. The channel is closed by the Disconnect function signaling the listeners to halt.
// This function is exposed to allow Driver implementations outside the nfcptl package.
func (c *Client) Terminate() <-chan struct{} {
	return c.terminate
}

// Commands returns the read only commands channel. The Driver MUST use this channel to listen for
// client commands and act accordingly.
// This function is exposed to allow Driver implementations outside the nfcptl package.
func (c *Client) Commands() <-chan Command {
	return c.commands
}

// SendCommand sends a ClientCommand to the internal commands channel. The Driver can act on
// this command when implemented by publishing an Event.
// When not implemented, the Driver should publish an error Event.
func (c *Client) SendCommand(cmd Command) {
	c.commands <- cmd
}

// VendorId returns the vendor ID the client is using.
func (c *Client) VendorId() uint16 {
	id, err := c.driver.VendorId(c.va)
	if err != nil {
		panic("nfcptl: invalid driver loaded")
	}
	return id
}

// ProductId returns the product ID the client is using.
func (c *Client) ProductId() uint16 {
	id, err := c.driver.ProductId(c.pa)
	if err != nil {
		panic("nfcptl: invalid driver loaded")
	}
	return id
}

// Debug returns whether the client is running in debug mode.
func (c *Client) Debug() bool {
	return c.debug
}
