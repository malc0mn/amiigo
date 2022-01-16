package amiibo

// Note that for these tests to succeed you will need to add these files to the testdata folder:
//   - key_retail.bin: a file containing concatenated unfixed-info.bin and locked-secret.bin files
//   - real_amiibo.bin: a real 540 byte NFC dump of an amiibo character
//   - plain_amiibo.bin: the decrypted version of real_amiibo.bin in NFC format, not in amiitool
//     format
//   - plain_amiibo_amiitool.bin: the decrypted version of real_amiibo.bin in amiitool format

import (
	"bytes"
	"encoding/hex"
	"testing"
)

const (
	testKeyRetail   = "key_retail.bin"
	testPlainAmiibo = "plain_amiibo.bin"
	testRealAmiibo  = "real_amiibo.bin"
)

func loadRetailKey(t *testing.T) *RetailKey {
	file := testDataDir + testKeyRetail
	key, err := NewRetailKey(file)
	if err != nil {
		t.Fatalf("NewRetailKey returned error %s. Make sure you have the correct %s file!", err, file)
	}
	return key
}

func loadRealAmiibo(t *testing.T, file string) *Amiibo {
	typ := "decrypted"
	if file == testRealAmiibo {
		typ = "real"
	}

	data := readFileWithError(t, file, "failed to load file %s, provide a "+typ+" amiibo dump for testing")
	amiibo, err := NewAmiibo(data, nil)
	if err != nil {
		t.Fatalf("NewAmiibo failed: got %s, want nil", err)
	}
	return amiibo
}

func TestNewRetailKey(t *testing.T) {
	wrong := []string{
		"non-existant.bin",
		"crypto_short_key_retail.bin",
		"crypto_long_key_retail.bin",
		"crypto_wrong_key_retail.bin",
	}

	for _, f := range wrong {
		key, err := NewRetailKey(testDataDir + f)
		if key != nil || err == nil {
			t.Fatalf("NewRetailKey should have failed, got %v, %s", key, err)
		}
	}

	key := loadRetailKey(t)

	want := "locked secret\000"
	got := key.Tag.TypeAsString()
	if got != want {
		t.Errorf("key.tag.TypeAsString got %s, want %s", got, want)
	}

	want = "unfixed infos\000"
	got = key.Data.TypeAsString()
	if got != want {
		t.Errorf("key.Data.TypeAsString got %s, want %s", got, want)
	}
}

func TestEncryptAmiibo(t *testing.T) {
	want := readFileWithError(t, testRealAmiibo, "failed to load file %s, provide a real amiibo dump for testing")

	got := Encrypt(loadRetailKey(t), loadRealAmiibo(t, testPlainAmiibo))
	if !bytes.Equal(got.Raw(), want) {
		t.Errorf("Encrypt got:\n%s want:\n%s", hex.Dump(got.Raw()), hex.Dump(want))
	}
}

func TestEncryptAmiitool(t *testing.T) {
	want := readFileWithError(t, testRealAmiibo, "failed to load file %s, provide a real amiibo dump for testing")
	data := readFileWithError(t, "plain_amiibo_amiitool.bin", "failed to load file %s, provide a decrypted amiibo dump for testing")

	amiitool, err := NewAmiitool(data, nil)
	if err != nil {
		t.Fatalf("NewAmiitool: got %s, want nil", err)
	}

	enc := Encrypt(loadRetailKey(t), amiitool)
	got, err := NewAmiibo(nil, enc.(*Amiitool))
	if !bytes.Equal(got.Raw(), want) {
		t.Errorf("Encrypt got:\n%s want:\n%s", hex.Dump(got.Raw()), hex.Dump(want))
	}
	if err != nil {
		t.Errorf("Encrypt got %s, want nil", err)
	}
}

func TestDecrypt(t *testing.T) {
	want := readFileWithError(t, testPlainAmiibo, "Encrypt failed to load file %s, provide a decrypted amiibo dump for testing")

	got, err := Decrypt(loadRetailKey(t), loadRealAmiibo(t, testRealAmiibo))
	if !bytes.Equal(got.Raw(), want) {
		t.Errorf("Decrypt got:\n%s want:\n%s", hex.Dump(got.Raw()), hex.Dump(want))
	}
	if err != nil {
		t.Errorf("Decrypt got %s, want nil", err)
	}
}

func TestDecryptFail(t *testing.T) {
	_, err := Decrypt(loadRetailKey(t), loadRealAmiibo(t, testDummyNtag))
	if err == nil {
		t.Error("Decrypt got nil, want error")
	}
}
