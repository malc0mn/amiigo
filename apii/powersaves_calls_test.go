package apii

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
)

func getSettings(t *testing.T) *Settings {
	file := "settings.xml"
	s, err := NewSettings(readFile(t, file))
	if err != nil {
		t.Fatalf("could not unmarshal file %s, error %s", file, err)
	}

	return s
}

func assertRequest(t *testing.T, r *http.Request, u, m string) {
	assertUrl(t, r, u)
	assertMethod(t, r, m)
}

func assertUrl(t *testing.T, r *http.Request, want string) {
	got := r.URL.String()
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func assertMethod(t *testing.T, r *http.Request, want string) {
	got := r.Method
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestPowerSavesAPI_GetSettings(t *testing.T) {
	a := NewPowerSavesAPI(newTestClient(func(req *http.Request) *http.Response {
		assertRequest(t, req, "http://settings.powersaves.net/psa/settings.xml", "GET")
		return &http.Response{
			StatusCode: 200,
			Body:       fromFile(t, "settings.xml"),
		}
	}))

	s, err := a.GetSettings()
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}
	if s == nil {
		t.Errorf("got %v, want Settings struct", s)
	}
}

func TestPowerSavesAPI_GetCodelist(t *testing.T) {
	a := NewPowerSavesAPI(newTestClient(func(req *http.Request) *http.Response {
		assertRequest(t, req, "http://settings.powersaves.net/psa/codelist.xml", "GET")
		return &http.Response{
			StatusCode: 200,
			Body:       fromFile(t, "codelist.xml"),
		}
	}))

	cl, err := a.GetCodelist(getSettings(t))
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}
	if cl == nil {
		t.Errorf("got %v, want CheatList struct", cl)
	}
}

func TestPowerSavesAPI_GetAuthorization(t *testing.T) {
	a := NewPowerSavesAPI(newTestClient(func(req *http.Request) *http.Response {
		assertRequest(t, req, "https://psaapp.powersaves.net/api/Authorisation", "GET")
		return &http.Response{
			StatusCode: 200,
			Body:       fromFile(t, "authorisation.xml"),
		}
	}))

	vf, err := a.GetAuthorization(getSettings(t))
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}
	if vf == nil {
		t.Errorf("got %v, want VerifyResponse struct", vf)
	}
}

func TestPowerSavesAPI_PostCheat(t *testing.T) {
	a := NewPowerSavesAPI(newTestClient(func(req *http.Request) *http.Response {
		assertRequest(t, req, "https://psaapp.powersaves.net/api/codes", "POST")
		got := req.Header.Get("Content-Type")
		want := "multipart/form-data; boundary="
		if !strings.HasPrefix(got, want) {
			t.Errorf("got %s, want %s", got, want)
		}

		req.ParseMultipartForm(0)

		log.Printf("%v", req.MultipartForm.File["Character"][0].Filename)

		type file struct {
			name string
			data []byte
		}

		wantf := map[string]file{
			"Character": {"character.bin", []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
			"Payload":   {"payload.bin", []byte{0x01, 0x02, 0x03}},
		}

		for k, f := range wantf {
			d := req.MultipartForm.File[k][0]
			gotn := d.Filename
			if gotn != f.name {
				t.Errorf("got %s, want %s", gotn, f.name)
			}
			file, err := d.Open()
			if err != nil {
				t.Fatalf("failed opening file %s", gotn)
			}
			gotd, err := io.ReadAll(file)
			if err != nil {
				t.Fatalf("failed reading file %s", gotn)
			}
			if !bytes.Equal(gotd, f.data) {
				t.Errorf("got %v, want %v", gotd, f.data)
			}
		}

		wantv := map[string]string{
			"Address":       "0x52",
			"PayloadLength": "0x03",
		}

		for field, want := range wantv {
			got := req.Form.Get(field)
			if got != want {
				t.Errorf("got %s, want %s", got, want)
			}
		}

		return &http.Response{
			StatusCode: 200,
			Body:       fromFile(t, "authorisation.xml"), // Return body not important in this test
		}
	}))

	b, err := a.PostCheat(
		getSettings(t),
		&VerifyResponse{
			Token:    "",
			Vuid:     []byte{0x00, 0x01},
			Password: "",
		},
		&ApplyCheat{
			Character:     []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			Payload:       []byte{0x01, 0x02, 0x03},
			Address:       "0x52",
			PayloadLength: "0x03",
		},
	)
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}
	if b == nil || len(b) == 0 {
		t.Errorf("got %v, want non empty byte slice", b)
	}
}
