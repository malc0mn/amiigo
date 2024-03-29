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
	"errors"
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
	MagicBytes     [MaxMagicByteSize]byte
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
func NewDerivedKey(key *MasterKey, amiibo Amiidump) *DerivedKey {
	seed := Seed(key, amiibo)

	buf := make([]byte, 2+len(seed)) // Add 2 bytes to store the counter.
	copy(buf[2:], seed)              // The pass counter is prepended, so keep the first 2 bytes free.
	h := hmac.New(sha256.New, key.HmacKey[:])
	var b []byte
	pass := 0
	for len(b) < 48 { // 48 = 3 * 16 bytes which is the size of DerivedKey, but we'll end up with more.
		binary.BigEndian.PutUint16(buf[:2], uint16(pass)) // Prepend counter.
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

// Encrypt signs and encrypts the given amiibo. It returns a NEW amiibo struct. The original struct
// remains unaltered.
func Encrypt(key *RetailKey, amiibo Amiidump) Amiidump {
	// First calculate signature from the unencrypted data. This signature is used to validate
	// the data has been decrypted properly.
	t := NewDerivedKey(&key.Tag, amiibo)
	d := NewDerivedKey(&key.Data, amiibo)

	tHmac := NewTagHmac(t, amiibo)
	dHmac := NewDataHmac(d, amiibo, tHmac)

	amiibo.SetTagHMAC(tHmac)
	amiibo.SetDataHMAC(dHmac)

	// Now actually encrypt.
	return Crypt(d, amiibo)
}

// Decrypt decrypts the given amiibo. It returns a NEW amiibo struct. The original struct remains
// unaltered.
// An error is returned if verification after decryption fails. You WILL receive a decrypted Amiibo
// struct even if an error occured but beware that it might not contain valid amiibo data.
func Decrypt(key *RetailKey, amiibo Amiidump) (Amiidump, error) {
	t := NewDerivedKey(&key.Tag, amiibo)
	d := NewDerivedKey(&key.Data, amiibo)

	dec := Crypt(d, amiibo)

	if !Verify(dec, t, d) {
		return dec, errors.New("amiibo: HMAC signatures do not match")
	}

	return dec, nil
}

// Seed generates the Seed needed to calculate a DerivedKey using the given MasterKey and data.
func Seed(key *MasterKey, amiibo Amiidump) []byte {
	var seed []byte

	// Create 16 magic bytes.
	magicBytes := [MaxMagicByteSize]byte{}
	// Start with bytes 0x11 and 0x12 from our amiibo data, leaving 14 zeroed bytes.
	copy(magicBytes[:], amiibo.WriteCounter())

	// Copy entire Type field.
	seed = append(seed, key.Type[:]...)
	// Append (MaxMagicByteSize - int(key.MagicBytesSize)) from the input Seed.
	seed = append(seed, magicBytes[:MaxMagicByteSize-int(key.MagicBytesSize)]...)
	// Append all bytes from magicBytes.
	seed = append(seed, key.MagicBytes[:int(key.MagicBytesSize)]...)
	// Append 8 bytes of the tag UID...
	fullUid := amiibo.FullUID()
	seed = append(seed, fullUid[:8]...)
	// ..twice.
	seed = append(seed, fullUid[:8]...)
	// Xor bytes 96-127 of amiibo data with AES XOR pad and append them.
	salt := amiibo.Salt()
	for i := 0; i < 32; i++ {
		seed = append(seed, salt[i]^key.XorPad[i])
	}

	if len(seed) > MaxSeedSize {
		panic(fmt.Sprintf("amiibo: Seed size %d larger than max %d", len(seed), MaxSeedSize))
	}

	return seed
}

// Crypt encrypts or decrypts the given data using the provided DerivedKey.
func Crypt(key *DerivedKey, amiibo Amiidump) Amiidump {
	block, err := aes.NewCipher(key.AesKey[:]) // 16 bytes key = AES-128
	if err != nil {
		panic("amiibo: unable to create AES cypher")
	}

	var dataIn []byte
	dataIn = append(dataIn, amiibo.RegisterInfoRaw()...)
	dataIn = append(dataIn, amiibo.SettingsRaw()...)
	dataOut := make([]byte, len(dataIn))

	stream := cipher.NewCTR(block, key.AesIV[:])
	stream.XORKeyStream(dataOut, dataIn)

	c, _ := NewAmiidump(amiibo.Raw(), amiibo.Type())
	c.SetRegisterInfo(dataOut[:32])
	c.SetSettings(dataOut[32:])

	return c
}

// NewTagHmac generates a new tag HMAC from the tag DerivedKey using unencrypted data.
func NewTagHmac(tagKey *DerivedKey, amiibo Amiidump) []byte {
	// Generate and tag HMAC.
	h := hmac.New(sha256.New, tagKey.HmacKey[:])
	fullUid := amiibo.FullUID()
	h.Write(fullUid[:8])
	modelSalt := amiibo.ModelInfoRaw()
	modelSalt = append(modelSalt, amiibo.Salt()...)
	h.Write(modelSalt)

	return h.Sum(nil)
}

// NewDataHmac generates a new data HMAC from the data DerivedKey using unencrypted data AND the
// tag HMAC generated by NewTagHmac.
func NewDataHmac(dataKey *DerivedKey, amiibo Amiidump, tagHmac []byte) []byte {
	// Generate and tag HMAC.
	h := hmac.New(sha256.New, dataKey.HmacKey[:])
	cntHmac := amiibo.WriteCounter()
	cntHmac = append(cntHmac, amiibo.Unknown2())
	cntHmac = append(cntHmac, amiibo.RegisterInfoRaw()...)
	h.Write(cntHmac)
	h.Write(amiibo.SettingsRaw())
	h.Write(tagHmac)
	fullUid := amiibo.FullUID()
	h.Write(fullUid[:8])
	modelSalt := amiibo.ModelInfoRaw()
	modelSalt = append(modelSalt, amiibo.Salt()...)
	h.Write(modelSalt)

	return h.Sum(nil)
}

// Verify checks if the decrypted data signature matches.
func Verify(amiibo Amiidump, tagKey, dataKey *DerivedKey) bool {
	// Generate HMACs from data and compare.
	tHmac := NewTagHmac(tagKey, amiibo)
	dHmac := NewDataHmac(dataKey, amiibo, tHmac)

	return hmac.Equal(amiibo.TagHMAC(), tHmac) && hmac.Equal(amiibo.DataHMAC(), dHmac)
}
