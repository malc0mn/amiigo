package apii

import (
	"encoding/json"
)

type AmiiboInfo struct {
	// AmiiboSeries holds the series the amiibo belongs to.
	AmiiboSeries string
	// Character holds the name of the amiibo character, the same character can have different amiibo designs.
	Character string
	// GameSeries holds the name of the game series the amiibo belongs to.
	GameSeries string
	// Games3DS holds a list of 3DS games it can be used in.
	Games3DS []GameInfo
	// GamesSwitch holds a list of Switch games it can be used in.
	GamesSwitch []GameInfo
	// GamesWiiU holds a list of Wii U games it can be used in.
	GamesWiiU []GameInfo
	// Head holds the first 8 bytes (0-7) of the amiibo ID represented as a hex string.
	Head string
	// Image holds the HTTP URL to the amiibo image file.
	Image string
	// Name of the amiibo character.
	Name string
	// The release date for North America, Japan, Europe and Australia.
	Release map[string]string
	// Tail holds the last 8 bytes (8-15) of the amiibo ID represented as a hex string.
	Tail string
	// Type holds the type it belongs to: card, figure or yarn.
	Type string
}

type GameInfo struct {
	GameID   []string
	GameName string
}

// NewAmiiboInfo creates a new AmiiboInfo struct given raw JSON data.
func NewAmiiboInfo(data []byte) (*AmiiboInfo, error) {
	ai := &struct {
		Amiibo *AmiiboInfo
	}{}
	if err := json.Unmarshal(data, ai); err != nil {
		return nil, err
	}

	return ai.Amiibo, nil
}

// NewAmiiboInfoList creates a new AmiiboInfo slice given raw JSON data.
func NewAmiiboInfoList(data []byte) ([]*AmiiboInfo, error) {
	ai := &struct {
		Amiibo []*AmiiboInfo
	}{}
	if err := json.Unmarshal(data, ai); err != nil {
		return nil, err
	}

	return ai.Amiibo, nil
}

// AmiiboInfoRequest is used to filter when querying for amiibo info. Fill in the fields you want
// the API to filter on and pass to GetAmiiboInfo.
type AmiiboInfoRequest struct {
	Name         string // Return the amiibo information base on its name.
	Id           string // Return the amiibo information base on its full id.
	Head         string // Return the amiibo information base on the first 8 bytes of the ID.
	Tail         string // Return the amiibo information base on the last 8 bytes of the ID.
	Type         string // Get all the amiibo based on its type as a list.
	Gameseries   string // Get all the amiibo based on the game series as a list.
	AmiiboSeries string // Get all the amiibo based on the series as a list.
	Character    string // Get all the amiibo based on its character as a list. Can also be a hex string like '0x1996'.
	Showgames    bool   // Also returns the games the amiibo can be used in.
	Showusage    bool   // Also returns the games the amiibo can be used in and how it's used in the game.
}

// KeyNameRequest is used to filter when querying for AmiiboSeries, GameSeries, Character or Type
// structs. Fill one of the fields you want to filter on and pass to the apropriate AmiiboAPI
// receiver.
type KeyNameRequest struct {
	Key  string
	Name string
}
