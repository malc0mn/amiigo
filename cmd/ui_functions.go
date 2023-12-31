package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/malc0mn/amiigo/amiibo"
	"github.com/malc0mn/amiigo/apii"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// showAmiiboInfo analyses the amiibo and updates the info, usage and image boxes.
func showAmiiboInfo(amb *amb, log, ifo, usg chan<- []byte, img *imageBox, baseUrl string) {
	if amb == nil || amb.a == nil {
		return
	}

	rawId := amb.a.ModelInfo().ID()
	if zeroed(rawId) {
		return
	}
	id := hex.EncodeToString(rawId)
	log <- encodeStringCell("Got id: " + id)

	typ := "a regular amiibo"
	if amb.a.Type() == amiibo.TypeAmiitool {
		typ = "an amiitool dump"
	}
	log <- encodeStringCell("Amiibo is " + typ)

	if amb.dec {
		log <- encodeStringCellWarning("Warning: amiibo is decrypted!")
	}

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
func loadDump(filename string, _ *amb, log chan<- []byte) bool {
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

	// TODO: how to detect amiitool files?
	am, err := amiibo.NewAmiibo(data, nil)
	if err != nil {
		log <- encodeStringCell(fmt.Sprintf("Error reading amiibo data: %s", err))
		return false
	}

	amiiboChan <- newAmiibo(am, false)

	log <- encodeStringCell("Amiibo read successful!")
	return true
}

// saveDump writes the active amiibo data to disk.
func saveDump(filename string, amb *amb, log chan<- []byte) bool {
	if amb == nil || amb.a == nil {
		log <- encodeStringCell("No amiibo data to write!")
		return false
	}
	if filename == "" {
		log <- encodeStringCell("Please provide a filename!")
		return false
	}

	filename = path.Clean(filename)

	ext := ".bin"
	filename = strings.TrimSuffix(filename, ext)
	if amb.dec {
		suf := "_decrypted"
		if !strings.HasSuffix(filename, suf) {
			ext = "_decrypted" + ext
		}
	}
	filename += ext

	dest := filename
	dir := filepath.Dir(filename)
	if dir == "." {
		dir, _ = os.Getwd()
		dest = filepath.Join(dir, filename)
	}

	log <- encodeStringCell(fmt.Sprintf("Writing amiibo to file '%s'", dest))
	if err := os.WriteFile(filename, amb.a.Raw(), 0644); err != nil {
		log <- encodeStringCell(fmt.Sprintf("Error writing file: %s", err))
		return false
	}

	log <- encodeStringCell("Amiibo dump successful!")
	return true
}

// prepData gets the amiibo data in the correct format for writing to the NFC portal.
// The returned byte array has the following structure:
//   - the first 8 bytes are the amiibo ID to be written
//   - the 9th byte is 0 for a full write and 1 to write only user data
//   - the rest of the bytes, 540 in total, is the amiibo data to be written
func prepData(value int, amb *amb, log chan<- []byte) []byte {
	if amb == nil || amb.a == nil {
		log <- encodeStringCell("Cannot write: please load amiibo data first!")
		return nil
	}

	if amb.dec {
		if !conf.expertMode {
			log <- encodeStringCellWarning("Refusing to write: decrypted amiibo!")
			return nil
		}
		log <- encodeStringCellWarning("WARNING: writing decrypted amiibo!")
	}

	var data []byte

	id := amb.a.ModelInfo().ID()

	switch amb.a.(type) {
	case *amiibo.Amiitool:
		data = amiibo.AmiitoolToAmiibo(amb.a.(*amiibo.Amiitool)).Raw()
	case *amiibo.Amiibo:
		data = amb.a.Raw()
	default:
		log <- encodeStringCell("Cannot write: unknown amiibo type!")
		return nil
	}

	return append(append(id, byte(value)), data...)
}

// writeToken will write the given data to the token on the NFC portal.
func writeToken(data []byte, nfcId []byte, ptl *portal, log chan<- []byte) {
	if data == nil || len(data) < 549 {
		return
	}

	user := data[8] == 1

	if user && nfcId != nil && !bytes.Equal(data[:8], nfcId) {
		if !conf.expertMode {
			log <- encodeStringCellWarning("Refusing to write: amiibo ID from dump does not match amiibo ID from portal!")
			return
		}
		log <- encodeStringCellWarning("WARNING: writing user data with mismatching amiibo ID!")
	}

	ptl.write(data[9:], user)
}

// decrypt decrypts the given amiibo and returns a new amiibo.Amiidump instance.
func decrypt(amb *amb, log chan<- []byte) *amb {
	if conf.retailKey == nil {
		log <- encodeStringCell("Cannot decrypt: no retail key loaded")
		return nil
	}
	if amb == nil || amb.a == nil {
		log <- encodeStringCell("Cannot decrypt: no amiibo data")
		return nil
	}
	dec, err := amiibo.Decrypt(conf.retailKey, amb.a)
	if err != nil {
		log <- encodeStringCell("Decryption error: " + err.Error())
		return nil
	}

	log <- encodeStringCell("Decryption successful")
	return newAmiibo(dec, amb.nfc)
}
