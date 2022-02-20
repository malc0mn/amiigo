package nfcptl

// Protocol represents the Protocol layer used by the client.
type Protocol interface {
	// Connect connects to the device.
	Connect(c *Client) error
	// Disconnect cleanly disconnects from the device.
	Disconnect() error
	// Read reads data from the underlying Protocol into the supplied buffer.
	Read(buf []byte) (int, error)
	// Write writes the given data to the underlying Protocol.
	Write(p []byte) (int, error)
}
