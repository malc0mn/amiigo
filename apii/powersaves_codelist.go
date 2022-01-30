package apii

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	PS_TypeList       = "List"
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
	Idx    int     `xml:"idx"` // An internal index
	Name   string  // The game name
	Folder *Folder `xml:"-"` // The folder containing all cheats
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

type Cheat struct {
	Idx      int `xml:"idx"`
	Name     string
	GamesIds GameIds `xml:"Games"`
	Desc     string
	Data     string
	Codes    []*Code `xml:"Codes>Code"`
}

type List struct {
	Idx    int      `xml:"idx"`
	Cheats []*Cheat `xml:"Entries>Cheat"`
}

type Code struct {
	Addr string
	Data string
}

// Address returns the string that should be posted as Address when applying the cheat.
func (co *Code) Address() string {
	return addrToHexString(co.Addr[2:])
}

type Folder struct {
	Name    string
	GameIds GameIds `xml:"Games"`
	Idx     string  `xml:"idx"`
	Addr    string
	Type    string   // The UI element to use for this folder
	Cheats  []*Cheat `xml:"Cheat"`
	ListIdx int      `xml:"List"`
	List    *List    `xml:"-"`
}

// Address returns the string that should be posted as Address when applying the cheat.
func (f *Folder) Address() string {
	if f.Addr == "" {
		return f.Addr
	}
	return addrToHexString(f.Addr[2:])
}

// SliderMinMax returns a slice of integers containing the min. and max. allowed value for a Cheat
// that is inside a PS_TypeSlider or PS_TypeSliderBig type of folder.
// Returns nil when the cheat is not found in the folder.
func (f *Folder) SliderMinMax(c *Cheat) []int {
	for _, fc := range f.Cheats {
		if c == fc && (f.Type == PS_TypeSlider || f.Type == PS_TypeSliderBig) {
			return slider(f.Cheats[0].Data)
		}
	}

	return nil
}

func (f *Folder) SliderVals(c *Cheat) map[int]int {
	for _, fc := range f.Cheats {
		if c == fc && (f.Type == PS_TypeSliderVals) {
			return sliderVals(f.Cheats[0].Data)
		}
	}

	return nil
}

// Payload returns the data that should be posted as Payload when applying the given Cheat.
func (f *Folder) Payload(c *Cheat, selection int) ([]byte, error) {
	for _, fc := range f.Cheats {
		if f.Type == PS_TypeList {
			for _, lc := range f.List.Cheats {
				if c == lc {
					if strings.Contains(lc.Data, " ") {
						// add leading zeros
					}
				}
			}
		}
		if c == fc {
			switch f.Type {
			case PS_TypeSliderVals:
				fallthrough
			case PS_TypeSliderBig:
				fallthrough
			case PS_TypeSlider:
				return int32ToByte(selection), nil
			case "Random":
				// TODO
			case "Code":
				// TODO
			}
		}
	}

	return nil, errors.New("unknown cheat")
}

type CheatList struct {
	// Games holds the list of games that have cheats.
	Games []*Game `xml:"Games>Game"`
	// GameIds is weird, not idea what use it has.
	GameIds GameIds `xml:"GameIds>Ids"`
	// Toys contains character IDs and their corresponding name. This is used to display the
	// character name for a toy placed on the portal.
	Toys []*Toy `xml:"Characters>Toy"`
	// Lists hold cheats that are attached to folders.
	Lists []*List `xml:"Lists>List"`
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

	for _, f := range cl.Folders {
		// Link lists to folders based on the list index and type.
		if f.Type == "List" {
			for _, l := range cl.Lists {
				if f.ListIdx == l.Idx {
					f.List = l
				}
			}
		}
		// Link folders to games based on the game index.
		for _, g := range cl.Games {
			for _, id := range f.GameIds {
				if id == g.Idx {
					g.Folder = f
				}
			}
		}
	}

	return cl, nil
}

func addrToHexString(a string) string {
	return "0x" + strings.TrimPrefix(a[2:], "0")
}

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

func hexTo32Bit(s string) []byte {
	src := []byte(fmt.Sprintf("%08s", s))

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
