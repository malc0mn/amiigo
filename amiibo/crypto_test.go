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
	"os"
	"testing"
)

const testDataDir = "testdata/"

func TestNewRetailKey(t *testing.T) {
	wrong := []string{
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

	file := testDataDir + "key_retail.bin"
	key, err := NewRetailKey(file)
	if err != nil {
		t.Fatalf("NewRetailKey returned error %s. Make sure you have the correct %s file!", err, file)
	}

	want := "locked secret\000"
	got := key.Tag.TypeAsString()
	if got != want {
		t.Errorf("key.tag.TypeAsString expected %s, got %s", want, got)
	}

	want = "unfixed infos\000"
	got = key.Data.TypeAsString()
	if got != want {
		t.Errorf("key.Data. expected %s, got %s", want, got)
	}
}

func TestEncryptAmiibo(t *testing.T) {
	file := testDataDir + "key_retail.bin"
	key, err := NewRetailKey(file)
	if err != nil {
		t.Fatalf("NewRetailKey returned error %s. Make sure you have the correct %s file!", err, file)
	}

	file = testDataDir + "real_amiibo.bin"
	want, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("EncryptAmiibo failed to load file %s, provide a real amiibo dump for testing", file)
	}

	file = testDataDir + "plain_amiibo.bin"
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("EncryptAmiibo failed to load file %s, provide a decrypted amiibo dump for testing", file)
	}

	amiibo, err := NewAmiibo(data, nil)
	if err != nil {
		t.Fatalf("NewAmiibo failed, expected nil, got %s", err)
	}

	got := Encrypt(key, amiibo)
	if !bytes.Equal(got.Raw(), want) {
		t.Errorf("EncryptAmiibo expected:\n%s got:\n%s", hex.Dump(want), hex.Dump(got.Raw()))
	}
}

func TestEncryptAmiitool(t *testing.T) {
	file := testDataDir + "key_retail.bin"
	key, err := NewRetailKey(file)
	if err != nil {
		t.Fatalf("NewRetailKey returned error %s. Make sure you have the correct %s file!", err, file)
	}

	file = testDataDir + "real_amiibo.bin"
	want, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("EncryptAmiitool failed to load file %s, provide a real amiibo dump for testing", file)
	}

	file = testDataDir + "plain_amiibo_amiitool.bin"
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("EncryptAmiitool failed to load file %s, provide a decrypted amiibo dump for testing", file)
	}

	amiitool, err := NewAmiitool(data, nil)
	if err != nil {
		t.Fatalf("NewAmiitool failed, expected nil, got %s", err)
	}

	enc := Encrypt(key, amiitool)
	got, _ := NewAmiibo(nil, enc.(*Amiitool))
	if !bytes.Equal(got.Raw(), want) {
		t.Errorf("EncryptAmiitool expected:\n%s got:\n%s", hex.Dump(want), hex.Dump(got.Raw()))
	}
}

func TestDecrypt(t *testing.T) {
	file := testDataDir + "key_retail.bin"
	key, err := NewRetailKey(file)
	if err != nil {
		t.Fatalf("NewRetailKey returned error %s. Make sure you have the correct %s file!", err, file)
	}

	file = testDataDir + "plain_amiibo.bin"
	want, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("Encrypt failed to load file %s, provide a decrypted amiibo dump for testing", file)
	}

	file = testDataDir + "real_amiibo.bin"
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("Encrypt failed to load file %s, provide a real amiibo dump for testing", file)
	}

	amiibo, err := NewAmiibo(data, nil)
	if err != nil {
		t.Fatalf("NewAmiibo failed, expected nil, got %s", err)
	}

	got, err := Decrypt(key, amiibo)
	if !bytes.Equal(got.Raw(), want) {
		t.Errorf("Decrypt expected:\n%s got:\n%s", hex.Dump(want), hex.Dump(got.Raw()))
	}
	if err != nil {
		t.Errorf("Decrypt expected nil got: %s", err)
	}
}
