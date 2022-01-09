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
	// DataHMAC returns the HMAC to be verified using a 'data' DerivedKey (master key unfixed-info.bin).
	DataHMAC() []byte
	FullUID() []byte
	Int() byte
	// ModelInfo returns a ModelInfo struct which can be used to extract detailed amiibo info. Since
	// this data is not encrypted, it can be accessed at any time.
	ModelInfo() *ModelInfo
	// ModelInfoRaw returns the raw amiibo model info.
	// The model info is also used in the calculation of the 'tag' HMAC concatenated with the Salt.
	ModelInfoRaw() []byte
	// Raw returns the raw tag data.
	Raw() []byte
	// RegisterInfo returns a RegisterInfo struct which can be used to extract detailed amiibo info.
	// This data is encrypted, so decrypt the amiibo first!
	RegisterInfo() *RegisterInfo
	// RegisterInfoRaw returns the first part of the data needed to generate the 'data' HMAC using the
	// DerivedKey generated from the 'data' master key (usually in a file named unfixed-info.bin).
	RegisterInfoRaw() []byte
	// Salt returns the 32 bytes used as salt in the Seed.
	Salt() []byte
	// SetDataHMAC sets the HMAC to sign the 'data' data.
	SetDataHMAC(dHmac []byte)
	SetRegisterInfo(enc []byte)
	SetSettings(enc []byte)
	// SetTagHMAC sets the HMAC to sign the 'tag' data.
	SetTagHMAC(tHmac []byte)
	Settings() *Settings
	// SettingsRaw returns the second block of crypto data. En/decryption must be done by
	// prepending RegisterInfoRaw and en/decrypting the entire block in one go.
	SettingsRaw() []byte
	StaticLockBytes() []byte
	// TagHMAC returns the HMAC to be verified using a 'tag' DerivedKey (master key locked-secret.bin).
	TagHMAC() []byte
	// Type returns the dump type: TypeAmiibo or TypeAmiitool.
	Type() DumpType
	// Unknown1 is obviously unknown but always seems to be set to 0xa5 which is done when writing to
	// the amiibo.
	Unknown1() byte
	// Unknown2 is obviously unknown but is used to generate the data HMAC.
	Unknown2() byte
	// WriteCounter returns the amiibo write counter. This counter is also used as magic bytes to
	// create the crypto Seed.
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
