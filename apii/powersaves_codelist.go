package apii

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

const (
	PS_TypeCode       = "Code"
	PS_TypeList       = "List"
	PS_TypeRandom     = "Random"
	PS_TypeSlider     = "Slider"
	PS_TypeSliderBig  = "SliderBig"
	PS_TypeSliderVals = "SliderVals"
)

type GameIds []int

// UnmarshalXML ensures a string of comma separated IDs is converted to an integer slice.
func (gi *GameIds) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var ids string
	if err := d.DecodeElement(&ids, &start); err != nil {
		return err
	}

	tmp := strings.Split(ids, ",")
	values := make([]int, len(tmp))
	for i, s := range tmp {
		v, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		values[i] = v
	}

	*gi = values

	return nil
}

// Game represents a game that has cheats.
type Game struct {
	Idx     int       `xml:"idx"` // An internal index
	Name    string    // The game name
	Folders []*Folder `xml:"-"` // The folders containing the cheats
}

// Toy represents an amiibo figure or card.
type Toy struct {
	Id   uint64 // The 64 bit amiibo ID
	Name string // The name of the character
}

// UnmarshalXML ensures proper conversion of the character IDs from hexadecimal to uint64.
func (t *Toy) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	tmp := &struct {
		Name string
		Idx  string `xml:"idx"`
	}{}
	if err := d.DecodeElement(tmp, &start); err != nil {
		return err
	}

	t.Name = tmp.Name
	var err error
	if t.Id, err = strconv.ParseUint(strings.Replace(tmp.Idx, "0x", "", 1), 16, 64); err != nil {
		return err
	}

	return nil
}

// Cheat represents a cheat that can be applied to a character.
type Cheat struct {
	Idx     int `xml:"idx"`
	Name    string
	GameIds GameIds `xml:"Games"`
	Desc    string
	Data    string
	// Codes holds the payload for a cheat. The XML seems to cater for multiple codes per cheat but
	// there is always just one code present. The array has been kept but currently only the first
	// code (position 0) is being used.
	Codes []*Code `xml:"Codes>Code"`
	// Folder is a backreference to the folder the cheat is in. If the cheat is in a folder, we'll
	// need to get the address from the folder to apply the cheat.
	Folder *Folder
}

// SliderMinMax returns a slice of integers containing the min. and max. allowed value for a Cheat
// that is inside a PS_TypeSlider or PS_TypeSliderBig folder.
// Returns nil when the cheat is in a different type of folder.
func (c *Cheat) SliderMinMax() []int {
	if c.Folder.IsTypeSlider() {
		return slider(c.Data)
	}

	return nil
}

// SliderVals returns a map of key value integers where the value will be the payload of the cheat.
// Returns nil when the cheat is in a different type of folder.
func (c *Cheat) SliderVals() map[int]int {
	if c.Folder.IsTypeSliderVals() {
		return sliderVals(c.Data)
	}

	return nil
}

// Payload returns the data that should be posted as Payload when applying the given Cheat.
// The selection parameter is required when there is a range of values to choose from. If there are
// no options to choose from, the parameter is ignored.
// Code cheats require multiple selection values which is why the
// Validating that the given selection is within the available range is the responsibility of the
// caller!
func (c *Cheat) Payload(selection ...int) ([]byte, error) {
	switch c.Folder.Type {
	case PS_TypeCode:
		return multiDataToBytes(selection), nil
	case PS_TypeList:
		return dataToBytes(c.Data), nil
	case PS_TypeRandom:
		return randomBytes(8)
	case PS_TypeSlider:
		fallthrough
	case PS_TypeSliderBig:
		fallthrough
	case PS_TypeSliderVals:
		return int32ToByte(selection[0]), nil
	}

	return nil, errors.New("unknown folder type")
}

// Address returns the address the cheat will be applied to. This address must be posted to the API
// to properly apply the cheat.
func (c *Cheat) Address() string {
	if len(c.Codes) > 0 {
		return c.Codes[0].Address() // Only the address of the first code must be posted!
	}

	return c.Folder.Address()
}

// List represents a list of cheats inside a folder.
type List struct {
	Idx    int     `xml:"idx"`
	Cheats []Cheat `xml:"Entries>Cheat"`
}

// Code represents one part of a cheat. If a cheat has multiple codes, each code must be posted to
// apply the cheat.
// Only one address, the address of the first code, must be posted in the Address field.
type Code struct {
	Addr string
	Data string
}

// Address returns the string that should be posted as Address when applying the cheat.
func (co *Code) Address() string {
	return addrToHexString(co.Addr)
}

// Folder represents a folder holding cheats.
type Folder struct {
	Name    string
	GameIds GameIds `xml:"Games"`
	Idx     string  `xml:"idx"`
	Addr    string
	Type    string   // The UI element to use for this folder
	Cheats  []*Cheat `xml:"Cheat"`
	ListIdx int      `xml:"List"`
}

// IsTypeSlider returns true if the folder is of type PS_TypeSlider or PS_TypeSliderBig.
func (f *Folder) IsTypeSlider() bool {
	return f.Type == PS_TypeSlider || f.Type == PS_TypeSliderBig
}

// IsTypeSliderVals returns true if the folder is of type PS_TypeSliderVals.
func (f *Folder) IsTypeSliderVals() bool {
	return f.Type == PS_TypeSliderVals
}

// Address returns the string that should be posted as Address when applying the cheat.
func (f *Folder) Address() string {
	if f.Addr == "" {
		return f.Addr
	}
	return addrToHexString(f.Addr)
}

// CheatList holds the full cheat list data returned by the API.
type CheatList struct {
	// Games holds the list of games that have cheats.
	Games []*Game `xml:"Games>Game"`
	// GameIds is weird, not idea what use it has.
	GameIds GameIds `xml:"GameIds>Ids"`
	// Toys contains character IDs and their corresponding name. This is used to display the
	// character name for a toy placed on the portal.
	Toys []*Toy `xml:"Characters>Toy"`
	// Lists hold cheats that are attached to folders.
	Lists []List `xml:"Lists>List"`
	// Folders hold lists of cheats that are attached to games.
	Folders []*Folder `xml:"Cheats>Folder"`
}

// CharacterNameById returns a character name for the given character ID. If the character is
// unknown, an empty string will be returned.
func (cl *CheatList) CharacterNameById(id uint64) string {
	for _, toy := range cl.Toys {
		if toy.Id == id {
			return toy.Name
		}
	}

	return ""
}

// NewCheatList creates a new CheatList struct given raw XML data.
func NewCheatList(data []byte) (*CheatList, error) {
	cl := &CheatList{}
	if err := xml.Unmarshal(data, cl); err != nil {
		return nil, err
	}

	if len(cl.Games) == 0 || len(cl.GameIds) == 0 || len(cl.Lists) == 0 || len(cl.Folders) == 0 || len(cl.Toys) == 0 {
		return nil, errors.New("unmarshal failed at least partially")
	}

	for _, f := range cl.Folders {
		// Copy list cheats to folder cheats based on the folder type and list index.
		if f.Type == PS_TypeList {
			for _, l := range cl.Lists {
				if f.ListIdx == l.Idx {
					f.Cheats = make([]*Cheat, len(l.Cheats))
					for i, c := range l.Cheats {
						cc := c
						f.Cheats[i] = &cc
					}
				}
			}
		}
		// Backlink cheats to folders to allow easy applying of cheats. Note that this MUST be
		// done AFTER the code block above!
		for _, c := range f.Cheats {
			c.Folder = f
		}
		// Link folders to games based on the game index.
		for _, g := range cl.Games {
			for _, id := range f.GameIds {
				if id == g.Idx {
					g.Folders = append(g.Folders, f)
				}
			}
		}
	}

	// Drop the now obsolete lists to free up memory.
	cl.Lists = nil

	return cl, nil
}

// TODO: we drop the first two bytes after investigating the post data, but what are they actually for?
//  They are 11, 12, 14 and 21.
func addrToHexString(a string) string {
	return "0x" + strings.TrimPrefix(a[2:], "0")
}

// sliderVals expects a space separated list of key value pairs, e.g. "key1 value1 key2 value2",
// that represent integers. Keys and values are in hexadecimal format.
// The string will be converted to a usable map.
func sliderVals(s string) map[int]int {
	p := strings.Split(s, " ")

	m := make(map[int]int, len(p)/2)

	var key int
	for i, v := range p {
		if i%2 == 0 {
			key = hexToInt(v)
		} else {
			m[key] = hexToInt(v)
		}
	}

	return m
}

// slider expects a space separated string containing a min. and max. value.
func slider(s string) []int {
	p := strings.Split(s, " ")

	min := hexTwosComplementToInt(p[0])
	max := hexToInt(p[1])

	if max+min != 0 {
		min = hexToInt(p[0])
	}

	return []int{min, max}
}

func hexTo8Bit(s string) []byte {
	return hexToXBit(s, 2)
}

func hexTo32Bit(s string) []byte {
	return hexToXBit(s, 8)
}

func hexToXBit(s string, precision int) []byte {
	src := []byte(fmt.Sprintf("%0"+strconv.Itoa(precision)+"s", s))

	dst := make([]byte, hex.DecodedLen(len(src)))
	_, err := hex.Decode(dst, src)
	if err != nil {
		panic(err)
	}

	return dst
}

func hexToInt(s string) int {
	return int(binary.BigEndian.Uint32(hexTo32Bit(s)))
}

func hexTwosComplementToInt(s string) int {
	return -int(^binary.BigEndian.Uint32(hexTo32Bit(s)) + 1)
}

func int32ToByte(i int) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, int32(i))
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// dataToBytes expects a space separated string of hexadecimal integers that will be converted to a
// byte array.
func dataToBytes(d string) []byte {
	r := strings.Split(d, " ")
	var b []byte
	for _, i := range r {
		b = append(b, hexTo8Bit(i)...)
	}

	return b
}

// TODO: this will need some efficiency improvement.
func multiDataToBytes(ints []int) []byte {
	var s []string
	for _, i := range ints {
		s = append(s, strconv.Itoa(i))
	}

	return dataToBytes(strings.Join(s, " "))
}

func randomBytes(size int) ([]byte, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	return b, nil
}
