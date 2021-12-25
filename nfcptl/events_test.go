package nfcptl

import "testing"

func TestString(t *testing.T) {
	list := map[EventType]string{
		DeviceName:   "DeviceName",
		HardwareInfo: "HardwareInfo",
		ApiPassword:  "ApiPassword",
	}

	for typ, want := range list {
		e := NewEvent(typ, []byte{})
		got := e.String()
		if want != got {
			t.Errorf("String() value was %s, want %s", got, want)
		}
	}
}
