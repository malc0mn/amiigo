package nfcptl

import (
	"github.com/pkg/term"
)

type Serial struct {
	Port string
	Baud int
}

// UART implements the Protocol interface to allow drivers to support UART devices.
type UART struct {
	prt *term.Term
}

func (uart *UART) Connect(c *Client) error {
	var err error

	setup, ok := c.Setup().(Serial)
	if !ok {
		panic("uart drivers must return a Serial struct")
	}

	if uart.prt, err = term.Open(setup.Port, term.Speed(setup.Baud), term.RawMode); err != nil {
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

func (uart *UART) Speed(s int) error {
	return uart.prt.SetSpeed(s)
}
