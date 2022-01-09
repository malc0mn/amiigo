package amiibo

import (
	"fmt"
	"testing"
)

func TestExtractBits(t *testing.T) {
	got := fmt.Sprintf("%032b", extractBits(1259835, 6, 3))
	want := "00000000000000000000000000100111"

	if got != want {
		t.Errorf("extractBits: got %s, want %s", got, want)
	}
}
