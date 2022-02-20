package nfcptl

import (
	"fmt"
	"github.com/google/gousb"
	"log"
	"time"
)

// DeviceSetup describes which config, interface, setting and in/out endpoints to use for a USB
// device.
type DeviceSetup struct {
	// Config holds the bConfigurationValue that needs to be set on the device for proper
	// initialisation. Most likely 1.
	Config int
	// Interface holds the bInterfaceNumber that needs to be used. Usually 0.
	Interface int
	// AlternateSetting holds the bAlternateSetting that needs to be used. Usually 0.
	AlternateSetting int
	// InEndpoint holds the device-to-host bEndpointAddress. In most cases this will be 1.
	InEndpoint int
	// OutEndpoint holds the host-to-device bEndpointAddress. In most cases this will be 1.
	OutEndpoint int
}

// USB implements the Protocol interface to allow drivers to support USB devices.
type USB struct {
	ctx   *gousb.Context     // The active context
	dev   *gousb.Device      // The active USB device
	iface *gousb.Interface   // The claimed USB interface
	cfg   *gousb.Config      // The active config
	in    *gousb.InEndpoint  // The device-to-host endpoint
	out   *gousb.OutEndpoint // The host-to-device endpoint
}

func (usb *USB) Connect(c *Client) error {
	var err error

	usb.ctx = gousb.NewContext()
	if c.Debug() {
		//usb.ctx.Debug(4)
	}

	usb.dev, err = usb.ctx.OpenDeviceWithVIDPID(gousb.ID(c.VendorId()), gousb.ID(c.ProductId()))
	if err != nil {
		return fmt.Errorf("could not open device: %v", err)
	}
	if usb.dev == nil {
		return fmt.Errorf("no device found for vid=%#04x,pid=%#04x", c.VendorId(), c.ProductId())
	}

	// AutoDetach is mandatory: it will detach the kernel driver before attempting to claim the
	// device. It will also reattach the kernel driver when we're done.
	usb.dev.SetAutoDetach(true)

	if c.Debug() {
		log.Printf("device: %s", usb.dev.String())
		log.Printf("devicedesc: %s", usb.dev.Desc.String())
		cdesc, _ := usb.dev.ConfigDescription(1)
		log.Printf("config: %v", cdesc)
	}

	setup, ok := c.Setup().(DeviceSetup)
	if !ok {
		panic("usb drivers must return a DeviceSetup struct")
	}

	usb.cfg, err = usb.dev.Config(setup.Config)
	if err != nil {
		return fmt.Errorf("failed to claim config %d of device %s: %v", setup.Config, usb.dev, err)
	}
	usb.iface, err = usb.cfg.Interface(setup.Interface, setup.AlternateSetting)
	if err != nil {
		return fmt.Errorf("failed to select interface #%d alternate setting %d of config %d of device %s: %v", setup.Interface, setup.AlternateSetting, setup.Config, usb.dev, err)
	}

	if usb.in, err = usb.iface.InEndpoint(setup.InEndpoint); err != nil {
		return fmt.Errorf("%s.InEndpoint(%d): %v", usb.iface, setup.InEndpoint, err)
	}

	if usb.out, err = usb.iface.OutEndpoint(setup.OutEndpoint); err != nil {
		return fmt.Errorf("%s.OutEndpoint(%d): %v", usb.iface, setup.OutEndpoint, err)
	}

	if c.Debug() {
		log.Println("Successfully connected!")
		log.Printf("Poll interval %d", usb.in.Desc.PollInterval)
		log.Printf("Max packet size %d", usb.in.Desc.MaxPacketSize)
	}

	return nil
}

func (usb *USB) Disconnect() error {
	if usb.iface != nil {
		usb.iface.Close()
	}

	if usb.cfg != nil {
		usb.cfg.Close()
	}

	if usb.dev != nil {
		usb.dev.Close()
	}

	if usb.ctx != nil {
		usb.ctx.Close()
	}

	return nil
}

func (usb *USB) Read(p []byte) (int, error) {
	return usb.in.Read(p)
}

func (usb *USB) Write(p []byte) (int, error) {
	return usb.out.Write(p)
}

// SetIdle sends SET_IDLE control request to the device.
// This function is exposed to allow Driver implementations outside the nfcptl package.
func (usb *USB) SetIdle(val, idx uint16) {
	usb.dev.Control(gousb.ControlOut|gousb.ControlClass|gousb.ControlInterface, bRequestSetIdle, val, idx, nil)
}

// PollInterval returns the poll interval for the NFC portal.
// This function is exposed to allow Driver implementations outside the nfcptl package.
func (usb *USB) PollInterval() time.Duration {
	return usb.in.Desc.PollInterval
}

// MaxPacketSize returns the maximum packet size the NFC portal will accept. Use a multiple of this
// value for optimal sending speed.
// This function is exposed to allow Driver implementations outside the nfcptl package.
func (usb *USB) MaxPacketSize() int {
	return usb.in.Desc.MaxPacketSize
}
