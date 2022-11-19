package apii

import "encoding/json"

type Type struct {
	Key  string
	Name string
}

// NewTypeList creates a new Type slice given raw JSON data.
func NewTypeList(data []byte) ([]Type, error) {
	ai := &struct {
		Amiibo []Type
	}{}
	if err := json.Unmarshal(data, ai); err != nil {
		return nil, err
	}

	return ai.Amiibo, nil
}
