package apii

import (
	"io"
	"net/http"
	"os"
	"testing"
)

const (
	testDataDir = "testdata/"
)

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func newTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{Transport: fn}
}

func fromFile(t *testing.T, file string) io.ReadCloser {
	fh, err := os.Open(testDataDir + file)
	if err != nil {
		t.Fatalf("failed to open file %s", file)
	}

	return fh
}

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
