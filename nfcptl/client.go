package nfcptl

import (
	"fmt"
	"github.com/google/gousb"
	"log"
	"time"
)

// Client allows easy communications with an NFC portal connected over USB.
type Client struct {
	driver Driver // The driver to use for communicating with the NFC portal

	debug bool // Will enable verbose logging

	ctx   *gousb.Context     // The active context
	dev   *gousb.Device      // The active USB device
	iface *gousb.Interface   // The claimed USB interface
	done  func()             // A cleanup function to close the interface and config
	in    *gousb.InEndpoint  // The device-to-host endpoint
	out   *gousb.OutEndpoint // The host-to-device endpoint
}

// NewClient builds a new Client struct.
func NewClient(vendor, device string, debug bool) (*Client, error) {
	d, err := GetDriverByVendorAndProductAlias(vendor, device)
	if err != nil {
		return nil, err
	}

	c := &Client{
		driver: d,
		debug:  debug,
	}

	if c.debug {
		log.Printf("Using vendor ID 0x%s and product ID 0x%s", c.driver.VendorId(), c.driver.ProductId())
	}

	return c, nil
}

// Connect establishes a new connection to the device and opens up input and output endpoints.
func (c *Client) Connect() error {
	var err error

	c.ctx = gousb.NewContext()
	if c.debug {
		c.ctx.Debug(4)
	}

	c.dev, err = c.ctx.OpenDeviceWithVIDPID(c.driver.VendorId(), c.driver.ProductId())
	if err != nil {
		return fmt.Errorf("could not open device: %v", err)
	}
	if c.dev == nil {
		return fmt.Errorf("no device found for vid=%s,pid=%s", c.driver.VendorId(), c.driver.ProductId())
	}

	// AutoDetach is mandatory: it will detach the kernel driver before attempting to claim the
	// device. It will also reattach the kernel driver when we're done.
	c.dev.SetAutoDetach(true)

	if c.debug {
		log.Printf("device: %s", c.dev.String())
		log.Printf("devicedesc: %s", c.dev.Desc.String())
		cdesc, _ := c.dev.ConfigDescription(1)
		log.Printf("config: %v", cdesc)
	}

	// Let's hope all NFC portals are plain and simple, otherwise we'll have to move some more stuff to the driver.
	if c.iface, c.done, err = c.dev.DefaultInterface(); err != nil {
		return fmt.Errorf("%s.DefaultInterface(): %v", c.dev, err)
	}

	if c.in, err = c.iface.InEndpoint(c.driver.InEndpoint()); err != nil {
		return fmt.Errorf("%s.InEndpoint(%d): %v", c.iface, c.driver.InEndpoint(), err)
	}

	if c.out, err = c.iface.OutEndpoint(c.driver.OutEndpoint()); err != nil {
		return fmt.Errorf("%s.OutEndpoint(%d): %v", c.iface, c.driver.OutEndpoint(), err)
	}

	if c.debug {
		log.Println("Successfully connected!")
		log.Printf("Poll interval %d", c.in.Desc.PollInterval)
		log.Printf("Max packet size %d", c.in.Desc.MaxPacketSize)
	}

	go c.Write(c.out.Desc.PollInterval, c.out.Desc.MaxPacketSize)
	go c.Read(c.in.Desc.PollInterval, c.in.Desc.MaxPacketSize)

	return nil
}

// Disconnect cleanly disconnects the client and frees up all resources.
func (c *Client) Disconnect() error {
	// Calling done() closes the interface and the config.
	if c.done != nil {
		c.done()
	}

	if c.dev != nil {
		c.dev.Close()
	}

	if c.ctx != nil {
		c.ctx.Close()
	}

	return nil
}

// Vid returns the vendor ID the client is using.
func (c *Client) Vid() gousb.ID {
	return c.driver.VendorId()
}

// Pid returns the product ID the client is using.
func (c *Client) Pid() gousb.ID {
	return c.driver.ProductId()
}

// Debug returns whether the client is running in debug mode.
func (c *Client) Debug() bool {
	return c.debug
}

// Read reads data from the NFC portal.
func (c *Client) Read(interval time.Duration, maxSize int) {
	c.driver.Read(c, interval, maxSize)
}

// Write sends data to the NFC portal.
func (c *Client) Write(interval time.Duration, maxSize int) {
	c.driver.Write(c, interval, maxSize)
}
