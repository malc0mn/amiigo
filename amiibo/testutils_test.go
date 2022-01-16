package amiibo

import (
	"os"
	"testing"
)

const (
	testDataDir = "testdata/"

	testDummyNtag    = "dummy_ntag215.bin"
	testDummyAmitool = "dummy_amiitool.bin"
)

func readFile(t *testing.T, fileName string) []byte {
	return readFileWithError(t, fileName, "failed to load file %s")
}

func readFileWithError(t *testing.T, fileName, error string) []byte {
	file := testDataDir + fileName
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf(error, file)
	}
	return data
}
