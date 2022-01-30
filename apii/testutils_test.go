package apii

import (
	"os"
	"testing"
)

const (
	testDataDir = "testdata/"
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

func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
