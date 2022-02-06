package apii

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

type PowerSavesAPI struct {
	client *http.Client
}

// GetSettings returns a new Settings struct which is needed to do any other API call. So this will
// be the first call to execute before doing any other call.
func (ps *PowerSavesAPI) GetSettings() (*Settings, error) {
	resp, err := ps.client.Get("http://settings.powersaves.net/psa/settings.xml")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return NewSettings(s)
}

// GetCodelist returns a new CheatList struct holding all games and their cheats as well as a list
// of known toys. You can get the name of a Toy by passing its ID to CheatList.CharacterNameById.
func (ps *PowerSavesAPI) GetCodelist(s *Settings) (*CheatList, error) {
	resp, err := ps.client.Get(s.CodelistUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	cl, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return NewCheatList(cl)
}

// GetAuthorization fetches a token and vuid from tha API. The vuid must be sent to the powersaves
// device which will return the corresponding password to be used together with the token as basic
// auth header in the PostCheat api call.
func (ps *PowerSavesAPI) GetAuthorization(s *Settings) (*VerifyResponse, error) {
	resp, err := ps.client.Get(s.AuthUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	vr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return NewVerifyResponse(vr)
}

func (ps *PowerSavesAPI) PostCheat(s *Settings, vr *VerifyResponse, ac *ApplyCheat) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	parts := []struct {
		field string
		file  string
		data  []byte
	}{
		{
			field: ac.CharacterFieldName(),
			file:  ac.CharacterFileName(),
			data:  ac.Character,
		},
		{
			field: ac.PayloadFieldName(),
			file:  ac.PayloadFileName(),
			data:  ac.Payload,
		},
		{
			field: ac.AddressFieldName(),
			file:  "",
			data:  []byte(ac.Address),
		},
		{
			field: ac.PayloadLengthFieldName(),
			file:  "",
			data:  []byte(ac.PayloadLength),
		},
	}

	for _, r := range parts {
		var (
			part io.Writer
			err  error
		)

		if r.file == "" {
			part, err = writer.CreateFormField(r.field)
		} else {
			part, err = writer.CreateFormFile(r.field, r.file)
		}
		if err != nil {
			return nil, err
		}

		part.Write(r.data)
	}

	writer.Close()

	req, err := http.NewRequest(http.MethodPost, s.ApplicationUrl, body)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(vr.Token, vr.Password)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	resp, err := ps.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// NewPowerSavesAPI returns a fresh PowerSavesAPI struct with the given client.
func NewPowerSavesAPI(client *http.Client) *PowerSavesAPI {
	return &PowerSavesAPI{client: client}
}
