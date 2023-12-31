package main

import "github.com/malc0mn/amiigo/amiibo"

// amb is an internal wrapper for amiibo data.
type amb struct {
	a   amiibo.Amiidump // The actual amiibo data.
	dec bool            // True when the amiibo is decrypted.
	nfc bool            // True when the amiibo was received from an NFC portal.
}

// newAmiibo creates a new amb struct.
func newAmiibo(a amiibo.Amiidump, nfc bool) *amb {
	return &amb{
		a:   a,
		dec: isAmiiboDecrypted(a, conf.retailKey),
		nfc: nfc,
	}
}

// isAmiiboDecrypted will TRY to ascertain if a given amiibo is decrypted. It will do this by
// assuming it is decrypted and verifying the HMAC signatures. If this fails, we assume it is
// encrypted.
// If no retail key is loaded, we will always assume it is not decrypted.
func isAmiiboDecrypted(a amiibo.Amiidump, key *amiibo.RetailKey) bool {
	if key == nil {
		return false
	}
	t := amiibo.NewDerivedKey(&key.Tag, a)
	d := amiibo.NewDerivedKey(&key.Data, a)

	return amiibo.Verify(a, t, d)
}
