package apii

import "encoding/json"

type Character struct {
	Key  string
	Name string
}

// NewCharacterList creates a new Character slice given raw JSON data.
func NewCharacterList(data []byte) ([]*Character, error) {
	ai := &struct {
		Amiibo []*Character
	}{}
	if err := json.Unmarshal(data, ai); err != nil {
		return nil, err
	}

	return ai.Amiibo, nil
}
