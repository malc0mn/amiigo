package main

import "testing"

func TestZeroed(t *testing.T) {
	tests := []struct {
		data []byte
		want bool
	}{
		{data: []byte{0, 0, 0, 0, 0, 0}, want: true},
		{data: []byte{0, 0, 1, 0, 0, 0}, want: false},
	}

	for _, test := range tests {
		got := zeroed(test.data)
		if got != test.want {
			t.Errorf("got %v, want %v", got, test.want)
		}
	}
}
