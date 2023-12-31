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

// isAmiiTool will TRY to ascertain if the given data is in the amiitool format. It will do this by
// assuming it is decrypted amiitool data. If that test fails, then it will assume encrypted
// amiitool data was given and try to decrypt it. If the decryption is successful, a pointer to an
// amiibo.Amiitool struct containing the original data is returned.
// So if it is valid amiitool data, you will always get a pointer to an amiibo.Amiitool struct
// holding the original data, nil otherwise.
// If no retail key is loaded, we will always assume it is not an amiitool format.
func isAmiiTool(data []byte, key *amiibo.RetailKey) *amiibo.Amiitool {
	if key == nil {
		return nil
	}

	a, err := amiibo.NewAmiitool(data, nil)
	if err != nil {
		return nil
	}

	if isAmiiboDecrypted(a, key) {
		return a
	}

	if _, err = amiibo.Decrypt(key, a); err != nil {
		return nil
	}

	return a
}
