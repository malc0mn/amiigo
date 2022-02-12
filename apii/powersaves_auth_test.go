package apii

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestSettings(t *testing.T) {
	file := "ps_settings.xml"
	s, err := NewSettings(readFile(t, file))
	if err != nil {
		t.Errorf("could not unmarshal file %s, error %s", file, err)
	}

	want := "http://settings.powersaves.net/psa/codelist.xml"
	got := s.CodelistUrl
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	want = "https://psaapp.powersaves.net/api/Authorisation"
	got = s.AuthUrl
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	want = "https://psaapp.powersaves.net/api/codes"
	got = s.ApplicationUrl
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	want = "http://www.codejunkies.com/powersaves-for-amiibo/"
	got = s.SoftwareUrl
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	want = "1.32"
	got = s.ClientVersion
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	wantb := true
	gotb := s.Active
	if gotb != wantb {
		t.Errorf("got %v, want %v", gotb, wantb)
	}
}

func TestAuthorisation(t *testing.T) {
	file := "ps_authorisation.xml"
	v, err := NewVerifyResponse(readFile(t, file))
	if err != nil {
		t.Errorf("could not unmarshal file %s, error %s", file, err)
	}

	want := "b91b27bc-ccb3-4d10-9ae3-497107a7a3fd"
	got := v.Token
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	vuid := "yx08a1IoH4D/sQZ1dil6cw=="
	wantv := make([]byte, len(vuid))
	base64.StdEncoding.Decode(wantv, []byte(vuid))
	gotv := v.Vuid
	if !bytes.Equal(gotv, wantv) {
		t.Errorf("got %s, want %s", gotv, wantv)
	}
}
