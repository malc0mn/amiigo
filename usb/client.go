package usb

import (
	"fmt"
	"github.com/google/gousb"
	"log"
)

// Client allows easy communications with an NFC portal connected over USB.
type Client struct {
	vid gousb.ID
	pid gousb.ID

	debug bool

	ctx   *gousb.Context
	dev   *gousb.Device
	iface *gousb.Interface
	done  func()
	in    *gousb.InEndpoint
	out   *gousb.OutEndpoint
}

// NewClient builds a new Client struct.
func NewClient(vendor, device string, debug bool) (*Client, error) {
	v, p, err := GetVidPidByVendorAndDeviceAlias(vendor, device)
	if err != nil {
		return nil, err
	}

	c := &Client{
		vid:   v,
		pid:   p,
		debug: debug,
	}

	if c.debug {
		log.Printf("Using vendor ID 0x%s and product ID 0x%s", c.vid, c.pid)
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

	c.dev, err = c.ctx.OpenDeviceWithVIDPID(c.vid, c.pid)
	if err != nil {
		return fmt.Errorf("could not open device: %v", err)
	}
	if c.dev == nil {
		return fmt.Errorf("no device found for vid=%s,pid=%s", c.vid, c.pid)
	}

	// AutoDetach is mandatory: it will detach the kernel driver before attempting to claim the device. It will also
	// reattach the kernel driver when we're done.
	c.dev.SetAutoDetach(true)

	if c.debug {
		log.Printf("device: %s", c.dev.String())
		log.Printf("devicedesc: %s", c.dev.Desc.String())
		cdesc, _ := c.dev.ConfigDescription(1)
		log.Printf("config: %v", cdesc)
	}

	if c.iface, c.done, err = c.dev.DefaultInterface(); err != nil {
		return fmt.Errorf("%s.DefaultInterface(): %v", c.dev, err)
	}

	// TODO: get IN endpoint from devices.go
	if c.in, err = c.iface.InEndpoint(0x81); err != nil {
		return fmt.Errorf("%s.OutEndpoint(0x81): %v", c.iface, err)
	}

	// TODO: get OUT endpoint from devices.go
	if c.out, err = c.iface.OutEndpoint(0x01); err != nil {
		return fmt.Errorf("%s.OutEndpoint(0x01): %v", c.iface, err)
	}

	if c.debug {
		log.Printf("Successfully connected!")
	}

	return nil
}

// Disconnect cleanly disconnects the client and frees up all resources.
func (c *Client) Disconnect() error {
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

// GetVid returns the vendor ID the client is using.
func (c *Client) GetVid() gousb.ID {
	return c.vid
}

// GetPid returns the product ID the client is using.
func (c *Client) GetPid() gousb.ID {
	return c.pid
}

// Debug returns whether the client is running in debug mode.
func (c *Client) Debug() bool {
	return c.debug
}
