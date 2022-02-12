package apii

import "encoding/json"

type GameSeries struct {
	Key  string
	Name string
}

// NewGameSeriesList creates a new GameSeries slice given raw JSON data.
func NewGameSeriesList(data []byte) ([]*GameSeries, error) {
	ai := &struct {
		Amiibo []*GameSeries
	}{}
	if err := json.Unmarshal(data, ai); err != nil {
		return nil, err
	}

	return ai.Amiibo, nil
}
