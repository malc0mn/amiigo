package amiibo

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"
)

const TestDataDir = "testdata/"

func TestNewRetailKey(t *testing.T) {
	wrong := []string{
		"crypto_short_key_retail.bin",
		"crypto_long_key_retail.bin",
		"crypto_wrong_key_retail.bin",
	}

	for _, f := range wrong {
		key, err := NewRetailKey(TestDataDir + f)
		if key != nil || err == nil {
			t.Fatalf("NewRetailKey should have failed, got %v, %s", key, err)
		}
	}

	file := TestDataDir + "key_retail.bin"
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

func TestEncrypt(t *testing.T) {
	file := TestDataDir + "key_retail.bin"
	key, err := NewRetailKey(file)
	if err != nil {
		t.Fatalf("NewRetailKey returned error %s. Make sure you have the correct %s file!", err, file)
	}

	file = TestDataDir + "real_amiibo.bin"
	want, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("Encrypt failed to load file %s, provide a real amiibo dump for testing", file)
	}

	file = TestDataDir + "plain_amiibo.bin"
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("Encrypt failed to load file %s, provide a decrypted amiibo dump for testing", file)
	}

	got := Encrypt(key, data)
	// TODO: fix this test since it will fail because of the new HMACs!
	if !bytes.Equal(got, want) {
		t.Errorf("Encrypt expected:\n%s got:\n%s", hex.Dump(want), hex.Dump(got[:]))
	}
}

func TestDecrypt(t *testing.T) {
	file := TestDataDir + "key_retail.bin"
	key, err := NewRetailKey(file)
	if err != nil {
		t.Fatalf("NewRetailKey returned error %s. Make sure you have the correct %s file!", err, file)
	}

	file = TestDataDir + "plain_amiibo.bin"
	want, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("Encrypt failed to load file %s, provide a decrypted amiibo dump for testing", file)
	}

	file = TestDataDir + "real_amiibo.bin"
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("Encrypt failed to load file %s, provide a real amiibo dump for testing", file)
	}

	got := Decrypt(key, data)
	if !bytes.Equal(got, want) {
		t.Errorf("Decrypt expected:\n%s got:\n%s", hex.Dump(want), hex.Dump(got[:]))
	}
}
