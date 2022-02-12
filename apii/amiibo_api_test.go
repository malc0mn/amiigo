package apii

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func emptyBody() io.ReadCloser {
	return io.NopCloser(strings.NewReader("{\"amiibo\":[]}"))
}

func TestAmiiboAPI_GetAmiiboInfo(t *testing.T) {
	a := NewAmiiboAPI(newTestClient(func(req *http.Request) *http.Response {
		assertRequest(t, req, "http://example.com/api/amiibo/", "GET")
		return &http.Response{
			StatusCode: 200,
			Body:       fromFile(t, "aa_amiibo_list.json"),
		}
	}), "http://example.com")

	air, err := a.GetAmiiboInfo(nil)
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}
	if air == nil {
		t.Errorf("got %v, want AmiiboList struct", air)
	}

	want := readFile(t, "aa_amiibo_list_processed.json")
	got, err := json.MarshalIndent(air, "", "  ")
	if err != nil {
		t.Fatal("unable to marshal data")
	}
	if !bytes.Equal(got, want) {
		t.Errorf("incorrect tree structure")
		fmt.Println("---------------------------------------------   got    ---------------------------------------------")
		fmt.Println(string(got))
		fmt.Println("--------------------------------------------- end got  ---------------------------------------------")
		fmt.Println("")
		fmt.Println("---------------------------------------------   want   ---------------------------------------------")
		fmt.Println(string(want))
		fmt.Println("--------------------------------------------- end want ---------------------------------------------")
	}
}

func TestAmiiboAPI_GetAmiiboInfoWitParams(t *testing.T) {
	a := NewAmiiboAPI(newTestClient(func(req *http.Request) *http.Response {
		assertRequest(t, req, "http://example.com/api/amiibo/?amiiboSeries=BoxBoy%21&character=0x1996&gameseries=Chibi+Robo&head=01010000&id=01010000000e0002&name=zelda&tail=000e0002&type=0x02&showgames&showusage", "GET")
		return &http.Response{
			StatusCode: 200,
			Body:       emptyBody(),
		}
	}), "http://example.com")

	air, err := a.GetAmiiboInfo(&AmiiboInfoRequest{
		Name:         "zelda",
		Id:           "01010000000e0002",
		Head:         "01010000",
		Tail:         "000e0002",
		Type:         "0x02",
		Gameseries:   "Chibi Robo",
		AmiiboSeries: "BoxBoy!",
		Character:    "0x1996",
		Showgames:    true,
		Showusage:    true,
	})
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}
	if air == nil {
		t.Errorf("got %v, want AmiiboList struct", air)
	}
}

func TestAmiiboAPI_GetAmiiboInfoWitBoolParams(t *testing.T) {
	a := NewAmiiboAPI(newTestClient(func(req *http.Request) *http.Response {
		assertRequest(t, req, "http://example.com/api/amiibo/?showgames&showusage", "GET")
		return &http.Response{
			StatusCode: 200,
			Body:       emptyBody(),
		}
	}), "http://example.com")

	air, err := a.GetAmiiboInfo(&AmiiboInfoRequest{
		Showgames: true,
		Showusage: true,
	})
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}
	if air == nil {
		t.Errorf("got %v, want AmiiboList struct", air)
	}
}

func TestLcFirst(t *testing.T) {
	want := "aLPHABET"
	got := lcFirst("ALPHABET")
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
