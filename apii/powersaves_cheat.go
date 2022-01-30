package apii

type ApplyCheat struct {
	Character     []byte // The full NTAG dump, send with filename "character.bin"
	Payload       []byte // The payload, send with filename "payload.bin"
	Address       []byte
	PayloadLength []byte
}
