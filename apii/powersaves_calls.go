package apii

import (
	"io"
	"io/ioutil"
	"net/http"
)

// GetSettings returns a new Settings struct which is needed to do any other API call. So this will
// be the first call to execute before doing any other call.
func GetSettings(h *http.Client) (*Settings, error) {
	resp, err := h.Get("http://settings.powersaves.net/psa/settings.xml")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return NewSettings(s)
}

// GetCodelist returns a new CheatList struct holding all games and their cheats as well as a list
// of known toys. You can get the name of a Toy by passing its ID to CheatList.CharacterNameById.
func GetCodelist(h *http.Client, s *Settings) (*CheatList, error) {
	resp, err := h.Get(s.CodelistUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cl, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return NewCheatList(cl)
}

func GetAuthorization(h *http.Client, s *Settings) (*VerifyResponse, error) {
	resp, err := h.Get(s.AuthUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	vr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return NewVerifyResponse(vr)
}

func PostCheat(h *http.Client, s *Settings, vr *VerifyResponse, ac *ApplyCheat) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, s.ApplicationUrl, http.NoBody)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(vr.Token, vr.Password)

	resp, err := h.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}