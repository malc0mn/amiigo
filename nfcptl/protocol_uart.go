package nfcptl

import (
	"github.com/tarm/serial"
)

// UART implements the Protocol interface to allow drivers to support UART devices.
type UART struct {
	prt *serial.Port
}

func (uart *UART) Connect(c *Client) error {
	var err error

	setup, ok := c.Setup().(serial.Config)
	if !ok {
		panic("uart drivers must return a serial.Config struct")
	}

	if uart.prt, err = serial.OpenPort(&setup); err != nil {
		return err
	}

	return nil
}

func (uart *UART) Disconnect() error {
	if uart.prt != nil {
		return uart.prt.Close()
	}

	return nil
}

func (uart *UART) Read(p []byte) (int, error) {
	return uart.prt.Read(p)
}

func (uart *UART) Write(p []byte) (int, error) {
	return uart.prt.Write(p)
}
