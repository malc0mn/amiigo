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
	return stringBody("{\"amiibo\":[]}")
}

func stringBody(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}

func assertResponse(t *testing.T, file string, res interface{}) {
	want := readFile(t, file)
	got, err := json.MarshalIndent(res, "", "  ")
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
func TestAmiiboAPI_GetAmiiboInfoByWrongId(t *testing.T) {
	a := NewAmiiboAPI(&http.Client{}, "http://example.com")

	tests := []string{
		"",
		"0",
		"1010000000e0002",   // too short
		"a1010000000e00025", // too long
		"01010000000g0002",  // wrong char lowercase
		"0G010000000e0002",  // wrong char uppercase
	}

	for _, id := range tests {
		_, err := a.GetAmiiboInfoById(id)
		want := "invalid id"
		if err == nil || err.Error() != want {
			t.Errorf("%s -- got %s, want %s", id, err, want)
		}
	}
}

func TestAmiiboAPI_GetAmiiboInfoById(t *testing.T) {
	a := NewAmiiboAPI(newTestClient(func(req *http.Request) *http.Response {
		assertRequest(t, req, "http://example.com/api/amiibo/?id=02c7000101220502", "GET")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       fromFile(t, "aa_amiibo_single.json"),
		}
	}), "http://example.com")

	ai, err := a.GetAmiiboInfoById("02c7000101220502")
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}
	if ai == nil {
		t.Errorf("got %v, want AmiiboInfo", ai)
	}

	assertResponse(t, "aa_amiibo_single_processed.json", ai)
}

func TestAmiiboAPI_GetAmiiboInfo(t *testing.T) {
	a := NewAmiiboAPI(newTestClient(func(req *http.Request) *http.Response {
		assertRequest(t, req, "http://example.com/api/amiibo/", "GET")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       fromFile(t, "aa_amiibo_list.json"),
		}
	}), "http://example.com")

	ai, err := a.GetAmiiboInfo(nil)
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}
	if ai == nil {
		t.Errorf("got %v, want AmiiboInfo slice", ai)
	}

	assertResponse(t, "aa_amiibo_list_processed.json", ai)
}

func TestAmiiboAPI_GetAmiiboInfoWithIdParam(t *testing.T) {
	a := NewAmiiboAPI(&http.Client{}, "http://example.com")

	_, err := a.GetAmiiboInfo(&AmiiboInfoRequest{Id: "01010000000e0002"})
	want := "use the GetAmiiboInfoById call to query by ID"
	if err == nil || err.Error() != want {
		t.Errorf("got %s, want %s", err, want)
	}
}

func TestAmiiboAPI_GetAmiiboInfoWithParams(t *testing.T) {
	a := NewAmiiboAPI(newTestClient(func(req *http.Request) *http.Response {
		assertRequest(t, req, "http://example.com/api/amiibo/?amiiboSeries=BoxBoy%21&character=0x1996&gameseries=Chibi+Robo&head=01010000&name=zelda&tail=000e0002&type=0x02&showgames&showusage", "GET")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       emptyBody(),
		}
	}), "http://example.com")

	ai, err := a.GetAmiiboInfo(&AmiiboInfoRequest{
		Name:         "zelda",
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
	if ai == nil {
		t.Errorf("got %v, want AmiiboInfo slice", ai)
	}
}

func TestAmiiboAPI_GetAmiiboInfoWithBoolParams(t *testing.T) {
	a := NewAmiiboAPI(newTestClient(func(req *http.Request) *http.Response {
		assertRequest(t, req, "http://example.com/api/amiibo/?showgames&showusage", "GET")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       emptyBody(),
		}
	}), "http://example.com")

	ai, err := a.GetAmiiboInfo(&AmiiboInfoRequest{
		Showgames: true,
		Showusage: true,
	})
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}
	if ai == nil {
		t.Errorf("got %v, want AmiiboInfo slice", ai)
	}
}

func TestAmiiboAPI_GetType(t *testing.T) {
	a := NewAmiiboAPI(newTestClient(func(req *http.Request) *http.Response {
		assertRequest(t, req, "http://example.com/api/type", "GET")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       fromFile(t, "aa_type_list.json"),
		}
	}), "http://example.com")

	typ, err := a.GetType(nil)
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}
	if typ == nil {
		t.Errorf("got %v, want Type slice", typ)
	}

	assertResponse(t, "aa_type_list_processed.json", typ)
}

func TestAmiiboAPI_GetTypeWithParams(t *testing.T) {
	a := NewAmiiboAPI(newTestClient(func(req *http.Request) *http.Response {
		assertRequest(t, req, "http://example.com/api/type?key=0x01", "GET")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       emptyBody(),
		}
	}), "http://example.com")

	typ, err := a.GetType(&KeyNameRequest{Key: "0x01", Name: "Card"})
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}
	if typ == nil {
		t.Errorf("got %v, want Type slice", typ)
	}
}

func TestAmiiboAPI_GetTypeWithNameParam(t *testing.T) {
	a := NewAmiiboAPI(newTestClient(func(req *http.Request) *http.Response {
		assertRequest(t, req, "http://example.com/api/type?name=Card", "GET")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       emptyBody(),
		}
	}), "http://example.com")

	typ, err := a.GetType(&KeyNameRequest{Name: "Card"})
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}
	if typ == nil {
		t.Errorf("got %v, want Type slice", typ)
	}
}

func TestAmiiboAPI_GetGameSeries(t *testing.T) {
	a := NewAmiiboAPI(newTestClient(func(req *http.Request) *http.Response {
		assertRequest(t, req, "http://example.com/api/gameseries", "GET")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       fromFile(t, "aa_gameseries_list.json"),
		}
	}), "http://example.com")

	gs, err := a.GetGameSeries(nil)
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}
	if gs == nil {
		t.Errorf("got %v, want GameSeries slice", gs)
	}

	assertResponse(t, "aa_gameseries_list_processed.json", gs)
}

func TestAmiiboAPI_GetAmiiboSeries(t *testing.T) {
	a := NewAmiiboAPI(newTestClient(func(req *http.Request) *http.Response {
		assertRequest(t, req, "http://example.com/api/amiiboseries", "GET")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       fromFile(t, "aa_amiiboseries_list.json"),
		}
	}), "http://example.com")

	as, err := a.GetAmiiboSeries(nil)
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}
	if as == nil {
		t.Errorf("got %v, want AmiiboSeries slice", as)
	}

	assertResponse(t, "aa_amiiboseries_list_processed.json", as)
}

func TestAmiiboAPI_GetCharacter(t *testing.T) {
	a := NewAmiiboAPI(newTestClient(func(req *http.Request) *http.Response {
		assertRequest(t, req, "http://example.com/api/character", "GET")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       fromFile(t, "aa_character_list.json"),
		}
	}), "http://example.com")

	char, err := a.GetCharacter(nil)
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}
	if char == nil {
		t.Errorf("got %v, want Character slice", char)
	}

	assertResponse(t, "aa_character_list_processed.json", char)
}

func TestAmiiboAPI_GetLastUpdated(t *testing.T) {
	a := NewAmiiboAPI(newTestClient(func(req *http.Request) *http.Response {
		assertRequest(t, req, "http://example.com/api/lastupdated", "GET")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       stringBody("{\"lastUpdated\": \"2019-03-18T16:34:10.688417\"}"),
		}
	}), "http://example.com")

	want := "2019-03-18T16:34:10.688417"
	got, err := a.GetLastUpdated()
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestLcFirst(t *testing.T) {
	want := "aLPHABET"
	got := lcFirst("ALPHABET")
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
