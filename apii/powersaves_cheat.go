package apii

import "fmt"

type ApplyCheat struct {
	Character     []byte // The full NTAG dump, send with filename "character.bin"
	Payload       []byte // The payload, send with filename "payload.bin"
	Address       string
	PayloadLength string
}

func (ac *ApplyCheat) CharacterFieldName() string {
	return "Character"
}

func (ac *ApplyCheat) CharacterFileName() string {
	return "character.bin"
}

func (ac *ApplyCheat) PayloadFieldName() string {
	return "Payload"
}

func (ac *ApplyCheat) PayloadFileName() string {
	return "payload.bin"
}

func (ac *ApplyCheat) AddressFieldName() string {
	return "Address"
}

func (ac *ApplyCheat) PayloadLengthFieldName() string {
	return "PayloadLength"
}

// NewApplyCheat returns an ApplyCheat struct for the given character data and Cheat ready for use
// with the PostCheat api call.
func NewApplyCheat(character []byte, cheat *Cheat, selection int) (*ApplyCheat, error) {
	pl, err := cheat.Payload(selection)
	if err != nil {
		return nil, err
	}

	return &ApplyCheat{
		Character:     character,
		Payload:       pl,
		Address:       cheat.Address(),
		PayloadLength: fmt.Sprintf("%#x", len(pl)),
	}, nil
}
