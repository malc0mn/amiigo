package apii

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"unicode"
)

type queryHandler func(interface{}, *http.Request)

type AmiiboAPI struct {
	client  *http.Client
	baseUrl string
}

// GetAmiiboInfo returns an AmiiboList struct enumerated with AmiiboInfo structs depending on the
// query sent to the API by means of the AmiiboInfoRequest struct.
// Pass nil if you want to get a full list.
func (aa *AmiiboAPI) GetAmiiboInfo(ar *AmiiboInfoRequest) ([]*AmiiboInfo, error) {
	b, err := aa.doGetRequest("/api/amiibo/", ar, addKeyValParams)
	if err != nil {
		return nil, err
	}

	return NewAmiiboInfoList(b)
}

func (aa *AmiiboAPI) GetType(kn *KeyNameRequest) ([]*Type, error) {
	b, err := aa.doGetRequest("/api/type", kn, addKeyNameFilter)
	if err != nil {
		return nil, err
	}

	return NewTypeList(b)
}

func (aa *AmiiboAPI) GetGameSeries(kn *KeyNameRequest) ([]*GameSeries, error) {
	b, err := aa.doGetRequest("/api/gameseries", kn, addKeyNameFilter)
	if err != nil {
		return nil, err
	}

	return NewGameSeriesList(b)
}

func (aa *AmiiboAPI) GetAmiiboSeries(kn *KeyNameRequest) ([]*AmiiboSeries, error) {
	b, err := aa.doGetRequest("/api/amiiboseries", kn, addKeyNameFilter)
	if err != nil {
		return nil, err
	}

	return NewAmiiboSeriesList(b)
}

func (aa *AmiiboAPI) GetCharacter(kn *KeyNameRequest) ([]*Character, error) {
	b, err := aa.doGetRequest("/api/character", kn, addKeyNameFilter)
	if err != nil {
		return nil, err
	}

	return NewCharacterList(b)
}

func (aa *AmiiboAPI) GetLastUpdated() (string, error) {
	resp, err := aa.client.Get(aa.baseUrl + "/api/lastupdated")
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	d := &struct {
		LastUpdated string
	}{}

	if err = json.Unmarshal(b, d); err != nil {
		return "", err
	}

	return d.LastUpdated, nil
}

func (aa *AmiiboAPI) doGetRequest(path string, q interface{}, qh queryHandler) ([]byte, error) {
	req, err := http.NewRequest("GET", aa.baseUrl+path, nil)
	if err != nil {
		return nil, err
	}

	if qh != nil {
		qh(q, req)
	}

	resp, err := aa.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// NewAmiiboAPI returns a fresh AmiiboAPI struct with the given client and base url.
func NewAmiiboAPI(client *http.Client, baseUrl string) *AmiiboAPI {
	return &AmiiboAPI{client: client, baseUrl: baseUrl}
}

// lcFirst converts the first character of a string to lower case.
func lcFirst(s string) string {
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

// addKeyNameFilter adds key OR name to query string. Key will take precedence!
func addKeyNameFilter(query interface{}, req *http.Request) {
	kn := query.(*KeyNameRequest)
	if kn == nil {
		return
	}

	if kn.Key != "" || kn.Name != "" {
		q := req.URL.Query()
		if kn.Key != "" {
			q.Add("key", kn.Key)
		} else if kn.Name != "" {
			q.Add("name", kn.Name)
		}
		req.URL.RawQuery = q.Encode()
	}
}

// addKeyValParams adds key value query parameters to the request.
func addKeyValParams(query interface{}, req *http.Request) {
	ar := query.(*AmiiboInfoRequest)
	if ar == nil {
		return
	}

	var bools []string
	q := req.URL.Query()
	v := reflect.Indirect(reflect.ValueOf(ar))
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		switch f.Kind() {
		case reflect.String:
			if f.String() != "" {
				q.Add(lcFirst(v.Type().Field(i).Name), f.String())
			}
		case reflect.Bool:
			if f.Bool() {
				bools = append(bools, lcFirst(v.Type().Field(i).Name))
			}
		}
	}
	if len(q) > 0 {
		req.URL.RawQuery = q.Encode()
	}
	// Weird in this API that booleans only have a 'key' and no value.
	if len(bools) > 0 {
		join := ""
		if req.URL.RawQuery != "" {
			join = "&"
		}
		req.URL.RawQuery += join + strings.Join(bools, "&")
	}
}