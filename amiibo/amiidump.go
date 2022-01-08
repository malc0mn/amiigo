package amiibo

import "fmt"

type DumpType byte

const (
	TypeAmiibo   DumpType = 1
	TypeAmiitool          = 2

	// AmiiboSize defines the minimum amount of bytes for an (incomplete) amiibo dump;
	AmiiboSize = 520
)

// Amiidump defines the interface necessary to encrypt and convert Amiibo and Amiitool structs.
type Amiidump interface {
	BCC1() byte
	CapabilityContainer() []byte
	CryptoSection() []byte
	DataHMAC() []byte
	DataHMACData1() []byte
	FullUID() []byte
	Int() byte
	ModelInfoRaw() []byte
	Nickname() string
	Raw() []byte
	Salt() []byte
	SetDataHMAC(dHmac []byte)
	SetEncrypt1(enc []byte)
	SetEncrypt2(enc []byte)
	SetTagHMAC(tHmac []byte)
	StaticLockBytes() []byte
	TagHMAC() []byte
	Type() DumpType
	Unknown() byte
	WriteCounter() []byte
}

// NewAmiidump creates a new Amiibo or Amiitool struct based on the given type.
func NewAmiidump(data []byte, typ DumpType) (Amiidump, error) {
	switch typ {
	case TypeAmiibo:
		return NewAmiibo(data, nil)
	case TypeAmiitool:
		return NewAmiitool(data, nil)
	}

	panic(fmt.Sprintf("amiibo: unknown dump type %d", typ))
}
