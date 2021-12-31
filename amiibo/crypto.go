package amiibo

// This file is derived from these source codes:
//  - https://github.com/socram8888/amiitool (c) 2015-2017 Marcos Del Sol Vives
//  - https://gist.github.com/anonymous/0a3e16f8f814deb2a056#file-amiibo-py-L126

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
)

const (
	KeyFileSize      = 160
	KeyFileMD5       = "45fd53569f5765eef9c337bd5172f937"
	KeyFileSha1      = "bbdbb49a917d14f7a997d327ba40d40c39e606ce"
	MaxMagicByteSize = 16
	MaxSeedSize      = 480
)

// MasterKey describes the structure of the info and secret keyfiles needed for amiibo crypto actions.
type MasterKey struct {
	HmacKey        [16]byte
	Type           [14]byte
	Rfu            byte
	MagicBytesSize byte
	MagicBytes     [16]byte
	XorPad         [32]byte
}

// TypeAsString returns the master key type as null terminated string.
func (mk *MasterKey) TypeAsString() string {
	return string(mk.Type[:])
}

// RetailKey describes the structure of the concatenated info and secret files.
type RetailKey struct {
	// Data holds the key usually named unfixed-info.bin
	Data MasterKey
	// Tag holds the key usually named locked-secret.bin
	Tag MasterKey
}

// NewRetailKey loads the key data from the given file and returns a new populated RetailKey
// struct.
func NewRetailKey(file string) (*RetailKey, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	if len(data) != KeyFileSize {
		return nil, fmt.Errorf("amiibo: invalid keyfile, expected %d bytes", KeyFileSize)
	}

	if fmt.Sprintf("%x", md5.Sum(data)) != KeyFileMD5 {
		return nil, fmt.Errorf("amiibo: invalid keyfile, expected md5 %s", KeyFileMD5)
	}

	if fmt.Sprintf("%x", sha1.Sum(data)) != KeyFileSha1 {
		return nil, fmt.Errorf("amiibo: invalid keyfile, expected sha1 %s", KeyFileSha1)
	}

	key := &RetailKey{}
	// Note that the byte order does not matter as we're using byte arrays.
	if err := binary.Read(bytes.NewReader(data), binary.BigEndian, key); err != nil {
		panic(fmt.Sprintf("amiibo: could not create new RetailKey %s", err))
	}

	if key.Tag.MagicBytesSize > MaxMagicByteSize || key.Data.MagicBytesSize > MaxMagicByteSize {
		return nil, fmt.Errorf("amiibo: magic byte size should not be larger than %d", MaxMagicByteSize)
	}

	return key, nil
}

// DerivedKey holds a derived key for a given amiibo figure.
type DerivedKey struct {
	AesKey  [16]byte
	AesIV   [16]byte
	HmacKey [16]byte
}

// NewDerivedKey is in essence a Deterministic Random Bit Generator that will generate a derived
// key from the given data.
func NewDerivedKey(key *MasterKey, data []byte) *DerivedKey {
	seed := Seed(key, data)

	buf := make([]byte, 2+len(seed)) // Add 2 bytes to store the counter.
	copy(buf[2:], seed)              // The pass counter is prepended, so keep the first 2 bytes free.
	h := hmac.New(sha256.New, key.HmacKey[:])
	var b []byte
	pass := 0
	for len(b) < 48 { // 48 = 3 * 16 bytes which is the size of DerivedKey, but we'll end up with more.
		binary.BigEndian.PutUint16(buf[0:2], uint16(pass)) // Prepend counter.
		if _, err := h.Write(buf); err != nil {
			panic("amiibo: could not hash buffer")
		}
		b = h.Sum(b)
		h.Reset()
		pass++
	}

	d := &DerivedKey{}
	// b is too long but binary.Read will stop when the struct is full, which is nice.
	// Note that the byte order does not matter as we're using byte arrays.
	if err := binary.Read(bytes.NewReader(b), binary.BigEndian, d); err != nil {
		panic("amiibo: could not populate new DerivedKey")
	}

	return d
}

// Encrypt signs and encrypts the given data.
func Encrypt(key *RetailKey, data []byte) []byte {
	// First calculate signature from the unencrypted data. This signature is used to validate
	// the data has been decrypted properly.
	t := NewDerivedKey(&key.Data, data)
	d := NewDerivedKey(&key.Data, data)

	tHmac := NewTagHmac(t, data)
	dHmac := NewDataHmac(d, data, tHmac)

	copy(data[52:84], tHmac)
	copy(data[128:160], dHmac)

	// Now actually encrypt.
	return Crypt(d, data)
}

// Decrypt decrypts the given data.
func Decrypt(key *RetailKey, data []byte) []byte {
	return Crypt(NewDerivedKey(&key.Data, data), data)
}

// Seed generates the seed needed to calculate a DerivedKey using the given MasterKey and data.
func Seed(key *MasterKey, data []byte) []byte {
	var seed []byte

	// Create 16 magic bytes.
	magicBytes := [MaxMagicByteSize]byte{}
	// Start with bytes 0x11 and 0x12 from our amiibo data, leaving 14 zeroed bytes.
	copy(magicBytes[:], data[17:19])

	// Copy entire Type field.
	seed = append(seed, key.Type[:]...)
	// Append (MaxMagicByteSize - int(key.MagicBytesSize)) from the input seed.
	seed = append(seed, magicBytes[:MaxMagicByteSize-int(key.MagicBytesSize)]...)
	// Append all bytes from magicBytes.
	seed = append(seed, key.MagicBytes[:int(key.MagicBytesSize)]...)
	// Append 8 bytes of the tag UID...
	seed = append(seed, data[0:8]...)
	// ..twice.
	seed = append(seed, data[0:8]...)
	// Xor bytes 96-127 of amiibo data with AES XOR pad and append them.
	for i := 0; i < 32; i++ {
		seed = append(seed, data[96+i]^key.XorPad[i])
	}

	if len(seed) > MaxSeedSize {
		panic(fmt.Sprintf("amiibo: seed size %d larger than max %d", len(seed), MaxSeedSize))
	}

	return seed
}

// Crypt encrypts or decrypts the given data using the provided DerivedKey.
func Crypt(key *DerivedKey, in []byte) []byte {
	block, err := aes.NewCipher(key.AesKey[:]) // 16 bytes key = AES-128
	if err != nil {
		panic("amiibo: unable to create AES cypher")
	}

	out := make([]byte, len(in))
	copy(out[:], in[:])

	var dataIn []byte
	dataIn = append(dataIn, in[20:52]...)
	dataIn = append(dataIn, in[160:520]...)
	dataOut := make([]byte, len(dataIn))

	stream := cipher.NewCTR(block, key.AesIV[:])
	stream.XORKeyStream(dataOut, dataIn)

	copy(out[20:52], dataOut[:32])
	copy(out[160:520], dataOut[32:])

	return out
}

// NewTagHmac generates a new tag HMAC from the tag DerivedKey using unencrypted data.
func NewTagHmac(key *DerivedKey, data []byte) []byte {
	// Generate and tag HMAC.
	h := hmac.New(sha256.New, key.HmacKey[:])
	h.Write(data[0:8])
	h.Write(data[84:128])

	return h.Sum(nil)
}

// NewDataHmac generates a new data HMAC from the data DerivedKey using unencrypted data AND the
// tag HMAC generated by NewTagHmac.
func NewDataHmac(key *DerivedKey, data, tagHmac []byte) []byte {
	// Generate and tag HMAC.
	h := hmac.New(sha256.New, key.HmacKey[:])
	h.Write(data[17:52])
	h.Write(data[160:520])
	h.Write(tagHmac)
	h.Write(data[0:8])
	h.Write(data[84:128])

	return h.Sum(nil)
}

// TODO: add Verify()!
