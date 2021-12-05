package nfcptl

import (
	"fmt"
	"github.com/google/gousb"
	"log"
)

// Client allows easy communications with an NFC portal connected over USB.
type Client struct {
	driver Driver // The driver to use for communicating with the NFC portal

	debug bool // Will enable verbose logging

	ctx   *gousb.Context     // The active context
	dev   *gousb.Device      // The active USB device
	iface *gousb.Interface   // The claimed USB interface
	cfg   *gousb.Config      // The active config
	in    *gousb.InEndpoint  // The device-to-host endpoint
	out   *gousb.OutEndpoint // The host-to-device endpoint

	events chan *Event // Device events will be received on this channel
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
		events: make(chan *Event, 10),
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

	setup := c.driver.Setup()

	c.cfg, err = c.dev.Config(setup.Config)
	if err != nil {
		return fmt.Errorf("failed to claim config %d of device %s: %v", setup.Config, c.dev, err)
	}
	c.iface, err = c.cfg.Interface(setup.Interface, setup.AlternateSetting)
	if err != nil {
		return fmt.Errorf("failed to select interface #%d alternate setting %d of config %d of device %s: %v", setup.Interface, setup.AlternateSetting, setup.Config, c.dev, err)
	}

	if c.in, err = c.iface.InEndpoint(setup.InEndpoint); err != nil {
		return fmt.Errorf("%s.InEndpoint(%d): %v", c.iface, setup.InEndpoint, err)
	}

	if c.out, err = c.iface.OutEndpoint(setup.OutEndpoint); err != nil {
		return fmt.Errorf("%s.OutEndpoint(%d): %v", c.iface, setup.OutEndpoint, err)
	}

	if c.debug {
		log.Println("Successfully connected!")
		log.Printf("Poll interval %d", c.in.Desc.PollInterval)
		log.Printf("Max packet size %d", c.in.Desc.MaxPacketSize)
	}

	// Pass control to the driver now.
	go c.driver.Drive(c, c.in.Desc.PollInterval, c.in.Desc.MaxPacketSize)

	return nil
}

// Disconnect cleanly disconnects the client and frees up all resources.
func (c *Client) Disconnect() error {
	if c.iface != nil {
		c.iface.Close()
	}

	if c.cfg != nil {
		c.cfg.Close()
	}

	if c.dev != nil {
		c.dev.Close()
	}

	if c.ctx != nil {
		c.ctx.Close()
	}

	close(c.events)

	return nil
}

// In returns the USB in endpoint for device-to-host communications.
func (c *Client) In() *gousb.InEndpoint {
	return c.in
}

// Out returns the USB out endpoint for host-to-device communications.
func (c *Client) Out() *gousb.OutEndpoint {
	return c.out
}

// Events returns the read only events channel. The caller can use this channel to listen for
// device events and act accordingly.
func (c *Client) Events() <-chan *Event {
	return c.events
}

// PublishEvent places an event on the event channel.
func (c *Client) PublishEvent(e *Event) {
	c.events <- e
}

// VendorId returns the vendor ID the client is using.
func (c *Client) VendorId() gousb.ID {
	return c.driver.VendorId()
}

// ProductId returns the product ID the client is using.
func (c *Client) ProductId() gousb.ID {
	return c.driver.ProductId()
}

// Debug returns whether the client is running in debug mode.
func (c *Client) Debug() bool {
	return c.debug
}

// SetIdle sends SET_IDLE control request to the device.
func (c *Client) SetIdle(val, idx uint16) {
	c.dev.Control(gousb.ControlOut|gousb.ControlClass|gousb.ControlInterface, bRequestSetIdle, val, idx, nil)
}
