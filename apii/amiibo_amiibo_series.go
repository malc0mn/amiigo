package apii

import "encoding/json"

type AmiiboSeries struct {
	Key  string
	Name string
}

// NewAmiiboSeriesList creates a new AmiiboSeries slice given raw JSON data.
func NewAmiiboSeriesList(data []byte) ([]*AmiiboSeries, error) {
	ai := &struct {
		Amiibo []*AmiiboSeries
	}{}
	if err := json.Unmarshal(data, ai); err != nil {
		return nil, err
	}

	return ai.Amiibo, nil
}
