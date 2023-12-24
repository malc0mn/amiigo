package main

import (
	"encoding/hex"
	"github.com/malc0mn/amiigo/amiibo"
	"github.com/malc0mn/amiigo/apii"
)

// showAmiiboInfo analyses the amiibo and updates the info, usage and image boxes.
func showAmiiboInfo(a *amiibo.Amiibo, log, ifo, usg chan<- []byte, img *imageBox, baseUrl string) {
	rawId := a.ModelInfo().ID()
	if zeroed(rawId) {
		return
	}
	id := hex.EncodeToString(rawId)
	log <- encodeStringCell("Got id: " + id)

	// Fill info box.
	log <- encodeStringCell("Fetching amiibo info")
	api := apii.NewAmiiboAPI(newCachedHttpClient(), baseUrl)
	ai, err := api.GetAmiiboInfoById(id)
	if err != nil {
		log <- encodeStringCell("API get amiibo info: " + err.Error())
		return
	}
	ifo <- formatAmiiboInfo(ai)

	// Fill image box.
	log <- encodeStringCell("Fetching image")
	i, err := getImage(ai.Image)
	if err != nil {
		log <- encodeStringCell("API get image: " + err.Error())
		return
	}
	img.setImage(i)

	// Fill usage box.
	log <- encodeStringCell("Fetching character usage")
	cu, err := api.GetCharacterUsage(ai.Character)
	if err != nil {
		log <- encodeStringCell("Api get character usage: " + err.Error())
		return
	}
	usg <- formatAmiiboUsage(cu, id)
}
