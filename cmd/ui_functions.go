package main

import (
	"encoding/hex"
	"fmt"
	"github.com/malc0mn/amiigo/amiibo"
	"github.com/malc0mn/amiigo/apii"
	"os"
	"path/filepath"
)

// showAmiiboInfo analyses the amiibo and updates the info, usage and image boxes.
func showAmiiboInfo(a amiibo.Amiidump, log, ifo, usg chan<- []byte, img *imageBox, baseUrl string) {
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

// loadDump loads an amiibo dump from disk.
func loadDump(filename string, _ amiibo.Amiidump, log chan<- []byte) bool {
	if filename == "" {
		log <- encodeStringCell("Please provide a filename!")
		return false
	}

	src := filename
	dir := filepath.Dir(filename)
	if dir == "." {
		dir, _ = os.Getwd()
		src = filepath.Join(dir, filename)
	}

	log <- encodeStringCell(fmt.Sprintf("Reading amiibo from file '%s'", src))

	data, err := os.ReadFile(filename)
	if err != nil {
		log <- encodeStringCell(fmt.Sprintf("Error reading file: %s", err))
		return false
	}

	am, err := amiibo.NewAmiibo(data, nil)
	if err != nil {
		log <- encodeStringCell(fmt.Sprintf("Error reading amiibo data: %s", err))
		return false
	}

	amiiboChan <- am

	log <- encodeStringCell("Amiibo read successful!")
	return true
}

// saveDump writes the active amiibo data to disk.
func saveDump(filename string, a amiibo.Amiidump, log chan<- []byte) bool {
	if a == nil {
		log <- encodeStringCell("No amiibo data to write!")
		return false
	}
	if filename == "" {
		log <- encodeStringCell("Please provide a filename!")
		return false
	}

	if ext := filepath.Ext(filename); ext != ".bin" {
		filename += ".bin"
	}

	dest := filename
	dir := filepath.Dir(filename)
	if dir == "." {
		dir, _ = os.Getwd()
		dest = filepath.Join(dir, filename)
	}

	log <- encodeStringCell(fmt.Sprintf("Writing amiibo to file '%s'", dest))
	if err := os.WriteFile(filename, a.Raw(), 0644); err != nil {
		log <- encodeStringCell(fmt.Sprintf("Error writing file: %s", err))
		return false
	}

	log <- encodeStringCell("Amiibo dump successful!")
	return true
}

// prepData gets the amiibo data in the correct format for writing to the NFC portal.
func prepData(a amiibo.Amiidump, dec bool, log chan<- []byte) []byte {
	if dec {
		log <- encodeStringCell("Refusing to write decrypted amiibo!")
		return nil
	}
	if a == nil {
		log <- encodeStringCell("Cannot write: please load amiibo data first!")
		return nil
	}

	switch a.(type) {
	case *amiibo.Amiitool:
		return amiibo.AmiitoolToAmiibo(a.(*amiibo.Amiitool)).Raw()
	case *amiibo.Amiibo:
		return a.Raw()
	default:
		panic(fmt.Sprintf("Unknown amiibo type!"))
	}
}

func decrypt(a amiibo.Amiidump, log chan<- []byte) amiibo.Amiidump {
	if a == nil {
		log <- encodeStringCell("Cannot decrypt: no amiibo data")
		return nil
	}
	if conf.retailKey == nil {
		log <- encodeStringCell("Cannot decrypt: no retail key loaded")
		return nil
	}
	dec, err := amiibo.Decrypt(conf.retailKey, a)
	if err != nil {
		log <- encodeStringCell("Decryption error: " + err.Error())
		return nil
	}

	log <- encodeStringCell("Decryption successful")
	return dec
}
