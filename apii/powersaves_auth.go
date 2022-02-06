package apii

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
)

// Settings holds the urls needed to communicate with the PowerSaves API.
type Settings struct {
	CodelistUrl    string `xml:"codelistUrl"`
	AuthUrl        string `xml:"authUrl"`
	ApplicationUrl string `xml:"applicationUrl"`
	SoftwareUrl    string `xml:"softwareUrl"`
	ClientVersion  string `xml:"clientVersion"`
	Active         bool   `xml:"active"`
}

// NewSettings creates a new Settings struct given raw XML data.
func NewSettings(data []byte) (*Settings, error) {
	s := &Settings{}
	if err := xml.Unmarshal(data, s); err != nil {
		return nil, err
	}

	if *s == (Settings{}) {
		return nil, errors.New("unmarshal resulted in empty struct")
	}

	return s, nil
}

// VerifyResponse holds a Token and Vuid as returned by the PowerSaves API. The Vuid must be sent
// to the PowerSaves NFC portal using the nfcptl.STM32F0_GenerateApiPassword command. The portal
// will respond with a password that can be used to construct a BasicAuth header as follows:
//   auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(Token+":"+Password))
// You should store the returned password in the VerifyResponse so it can be used to execute the
// PostCheat call.
type VerifyResponse struct {
	Token    string // Returned by the PowerSaves API
	Vuid     []byte // Returned by the PowerSaves API
	Password string `xml:"-"` // To be obtained from the NFC portal using the Vuid
}

// UnmarshalXML will take care of decoding the Vuid so that it is ready to send to the PowerSaves
// NFC portal.
func (vr *VerifyResponse) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	tmp := &struct {
		Token string
		Vuid  string
	}{}
	if err := d.DecodeElement(&tmp, &start); err != nil {
		return err
	}

	vr.Token = tmp.Token
	vr.Vuid = make([]byte, len(tmp.Vuid))
	if _, err := base64.StdEncoding.Decode(vr.Vuid, []byte(tmp.Vuid)); err != nil {
		return err
	}

	return nil
}

// NewVerifyResponse creates a new VerifyResponse struct given raw XML data.
func NewVerifyResponse(data []byte) (*VerifyResponse, error) {
	vr := &VerifyResponse{}
	if err := xml.Unmarshal(data, vr); err != nil {
		return nil, err
	}

	if vr.Token == "" || vr.Vuid == nil {
		return nil, errors.New("unmarshal failed at least partially")
	}

	return vr, nil
}
