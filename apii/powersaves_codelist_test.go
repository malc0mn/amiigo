package apii

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"reflect"
	"testing"
)

func TestCharacterNameById(t *testing.T) {
	file := "codelist.xml"
	c, err := NewCheatList(readFile(t, file))
	if err != nil {
		t.Errorf("could not unmarshal file %s, error %s", file, err)
	}

	want := "SSB - Fox"
	got := c.CharacterNameById(0x0580000000050002)

	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestNewCheatListTree(t *testing.T) {
	file := "codelist.xml"
	cl, err := NewCheatList(readFile(t, file))
	if err != nil {
		t.Fatalf("could not unmarshal file %s, error %s", file, err)
	}

	file = "tree.txt"
	want := readFile(t, file)

	var got bytes.Buffer
	for _, g := range cl.Games {
		got.WriteString("game: " + g.Name + "\n")
		for _, f := range g.Folders {
			got.WriteString("+-- folder: " + f.Name + "\n")
			got.WriteString("    type: " + f.Type + "\n")
			got.WriteString("    address: " + f.Address() + "\n")
			for _, c := range f.Cheats {
				got.WriteString("    +-- cheat: " + c.Name + "\n")
				got.WriteString("        address: " + c.Address() + "\n")
			}
		}
	}

	if !bytes.Equal(got.Bytes(), want) {
		t.Errorf("incorrect tree structure")
		fmt.Println("---------------------------------------------   got    ---------------------------------------------")
		fmt.Println(got.String())
		fmt.Println("--------------------------------------------- end got  ---------------------------------------------")
		fmt.Println("")
		fmt.Println("---------------------------------------------   want   ---------------------------------------------")
		fmt.Println(string(want))
		fmt.Println("--------------------------------------------- end want ---------------------------------------------")
	}
}

func TestLists(t *testing.T) {
	file := "codelist.xml"
	cl := &CheatList{}
	if err := xml.Unmarshal(readFile(t, file), cl); err != nil {
		t.Fatalf("could not unmarshal file %s, error %s", file, err)
	}

	wanti := 4
	goti := len(cl.Lists)
	if goti != wanti {
		t.Errorf("len(cl.Lists) got %d, want %d", goti, wanti)
	}

	wanti = 2
	goti = cl.Lists[2].Idx
	if goti != wanti {
		t.Errorf("cl.Lists.Id got %d, want %d", goti, wanti)
	}

	wanti = 34
	goti = len(cl.Lists[2].Cheats)
	if goti != wanti {
		t.Errorf("cl.Lists.Cheats got %d, want %d", goti, wanti)
	}

	wants := "Agility (Can be used in Four slots Max)"
	gots := cl.Lists[3].Cheats[5].Name
	if gots != wants {
		t.Errorf("cl.Lists.Cheats.Name got %s, want %s", gots, wants)
	}

	wants = ""
	gots = cl.Lists[3].Cheats[5].Desc
	if gots != wants {
		t.Errorf("cl.Lists.Cheats.Desc got %s, want %s", gots, wants)
	}

	wanti = 0
	goti = cl.Lists[3].Cheats[5].Idx
	if goti != wanti {
		t.Errorf("cl.Lists.Cheats.Id got %d, want %d", goti, wanti)
	}

	wants = "6"
	gots = cl.Lists[3].Cheats[5].Data
	if gots != wants {
		t.Errorf("cl.Lists.Cheats.Data got %s, want %s", gots, wants)
	}
}

func TestNewCheatList(t *testing.T) {
	file := "codelist.xml"
	cl, err := NewCheatList(readFile(t, file))
	if err != nil {
		t.Fatalf("could not unmarshal file %s, error %s", file, err)
	}

	wanti := 9
	goti := len(cl.Games)
	if goti != wanti {
		t.Errorf("len(cl.Games) got %d, want %d", goti, wanti)
	}

	wants := "Shovel Knight (3DS)"
	gots := cl.Games[5].Name
	if gots != wants {
		t.Errorf("cl.Games.Name got %s, want %s", gots, wants)
	}

	wanti = 8
	goti = cl.Games[5].Idx
	if goti != wanti {
		t.Errorf("cl.Games.Id got %d, want %d", goti, wanti)
	}

	wanti = 1
	goti = len(cl.Games[5].Folders[0].Cheats)
	if goti != wanti {
		t.Errorf("len(cl.Games.Folders.Cheats) got %d, want %d", goti, wanti)
	}

	wantsl := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 98, 99}
	gotsl := cl.GameIds
	if !equal(gotsl, wantsl) {
		t.Errorf("cl.GameIds got %v, want %v", gotsl, wantsl)
	}

	wanti = 699
	goti = len(cl.Toys)
	if goti != wanti {
		t.Errorf("len(cl.Toys) got %d, want %d", goti, wanti)
	}

	wants = "MSSS - Rosalina (Golf)"
	gots = cl.Toys[186].Name
	if gots != wants {
		t.Errorf("cl.Toys.Name got %s, want %s", gots, wants)
	}

	wanti = 0x09cf040102b70e02
	goti = int(cl.Toys[186].Id)
	if goti != wanti {
		t.Errorf("cl.Toys.Id got %#016x, want %#016x", goti, wanti)
	}

	wanti = 0
	goti = len(cl.Lists)
	if goti != wanti {
		t.Errorf("len(cl.Lists) got %d, want %d", goti, wanti)
	}

	wanti = 26
	goti = len(cl.Folders)
	if goti != wanti {
		t.Errorf("len(cl.Folders) got %d, want %d", goti, wanti)
	}

	wants = "Special Moves"
	gots = cl.Folders[4].Name
	if gots != wants {
		t.Errorf("cl.Folders.Name got %s, want %s", gots, wants)
	}

	wantsl = []int{0, 2}
	gotsl = cl.Folders[4].GameIds
	if !equal(gotsl, wantsl) {
		t.Errorf("cl.Folders.GameIds got %v, want %v", gotsl, wantsl)
	}

	wants = "f004"
	gots = cl.Folders[4].Idx
	if gots != wants {
		t.Errorf("cl.Folders.Id got %s, want %s", gots, wants)
	}

	wants = "Slider"
	gots = cl.Folders[4].Type
	if gots != wants {
		t.Errorf("cl.Folders.Type got %s, want %s", gots, wants)
	}

	wanti = 4
	goti = len(cl.Folders[4].Cheats)
	if goti != wanti {
		t.Errorf("cl.Folders.Folders got %d, want %d", goti, wanti)
	}

	if cl.Folders[4].Cheats[2].Folder != cl.Folders[4] {
		t.Error("cheat folder backreference does not match the folder it is in")
	}

	wants = "Up"
	gots = cl.Folders[4].Cheats[2].Name
	if gots != wants {
		t.Errorf("cl.Folders.Folders.Name got %s, want %s", gots, wants)
	}

	wantsl = []int{0, 2}
	gotsl = cl.Folders[4].Cheats[2].GameIds
	if !equal(gotsl, wantsl) {
		t.Errorf("cl.Folders.Folders.GameIds got %v, want %v", gotsl, wantsl)
	}

	wanti = 0
	goti = cl.Folders[4].Cheats[2].Idx
	if goti != wanti {
		t.Errorf("cl.Folders.Folders.Id got %d, want %d", goti, wanti)
	}

	wanti = 1
	goti = len(cl.Folders[4].Cheats[2].Codes)
	if goti != wanti {
		t.Errorf("cl.Folders.Folders.Codes got %d, want %d", goti, wanti)
	}

	wants = "110e7"
	gots = cl.Folders[4].Cheats[2].Codes[0].Addr
	if gots != wants {
		t.Errorf("cl.Folders.Folders.Codes.Addr got %s, want %s", gots, wants)
	}

	wants = "1 3"
	gots = cl.Folders[4].Cheats[2].Codes[0].Data
	if gots != wants {
		t.Errorf("cl.Folders.Folders.Codes.Data got %s, want %s", gots, wants)
	}

	wanti = 3
	goti = cl.Folders[14].ListIdx
	if goti != wanti {
		t.Errorf("cl.Folders.ListIdx got %d, want %d", goti, wanti)
	}
	wants = "210e1"
	gots = cl.Folders[14].Addr
	if gots != wants {
		t.Errorf("cl.Folders.Addr got %s, want %s", gots, wants)
	}

	wanti = 8
	goti = len(cl.Folders[14].Cheats)
	if goti != wanti {
		t.Errorf("cl.Folders.List got %d, want %d", goti, wanti)
	}
}

func TestCheatNoCodes(t *testing.T) {
	file := "codelist.xml"
	cl, err := NewCheatList(readFile(t, file))
	if err != nil {
		t.Fatalf("could not unmarshal file %s, error %s", file, err)
	}

	g := cl.Games[1]

	want := "Mario Party 10"
	got := g.Name
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}

	f := g.Folders[0]

	want = "Base Codes"
	got = f.Name
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}

	c := f.Cheats[8]
	want = "Watermelon Base Selected"
	got = c.Name
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}

	gotf := c.Folder
	if gotf != f {
		t.Errorf("got %v, want %v", gotf, f)
	}

	want = "0x109"
	got = c.Address()
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	wantsl := []byte{
		0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x05, 0x01,
		0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
	}
	gotsl, err := c.Payload(0)
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}
	if !bytes.Equal(gotsl, wantsl) {
		t.Errorf("got %#x, want %#x", gotsl, wantsl)
	}
}

func TestCheatWithCodes(t *testing.T) {
	file := "codelist.xml"
	cl, err := NewCheatList(readFile(t, file))
	if err != nil {
		t.Fatalf("could not unmarshal file %s, error %s", file, err)
	}

	g := cl.Games[2]

	want := "Super Smash Bros. (3DS)"
	got := g.Name
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}

	f := g.Folders[4]

	want = "Special Moves"
	got = f.Name
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}

	c := f.Cheats[2]
	want = "Up"
	got = c.Name
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}

	gotf := c.Folder
	if gotf != f {
		t.Errorf("got %v, want %v", gotf, f)
	}

	want = "0xe7"
	got = c.Address()
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	wantsl := []byte{0x00, 0x00, 0x00, 0x02}
	gotsl, err := c.Payload(2)
	if err != nil {
		t.Errorf("got %s, want nil", err)
	}
	if !bytes.Equal(gotsl, wantsl) {
		t.Errorf("got %#x, want %#x", gotsl, wantsl)
	}
}

func TestAddrToHexString(t *testing.T) {
	want := "0x109"
	got := addrToHexString("21109")
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	want = "0xdf"
	got = addrToHexString("210df")
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestSliderVals(t *testing.T) {
	want := map[int]int{
		1: 4, 2: 8, 3: 12, 4: 16, 5: 20, 6: 24, 7: 28, 8: 32, 9: 36, 10: 40, 11: 44,
		12: 48, 13: 52, 14: 56, 15: 60, 16: 64, 17: 68, 18: 72, 19: 76, 20: 80,
	}
	got := sliderVals("1 4 2 8 3 c 4 10 5 14 6 18 7 1c 8 20 9 24 a 28 b 2c c 30 d 34 e 38 f 3c 10 40 11 44 12 48 13 4c 14 50")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSlider(t *testing.T) {
	tests := map[string][]int{
		"1 3":         {1, 3},
		"1 8":         {1, 8},
		"0 270e":      {0, 9998},
		"ffffff38 c8": {-200, 200},
	}

	for s, want := range tests {
		got := slider(s)
		if !equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}

func TestHexToInt(t *testing.T) {
	want := 200
	got := hexToInt("c8")
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestHexTwosComplementToInt(t *testing.T) {
	want := -200
	got := hexTwosComplementToInt("ffffff38")
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestInt32ToByte(t *testing.T) {
	want := []byte{0xff, 0xff, 0xff, 0x3a}
	got := int32ToByte(-198)
	if !bytes.Equal(got, want) {
		t.Errorf("got %#x, want %#x", got, want)
	}
}

func TestDataToBytes(t *testing.T) {
	want := []byte{0x4e}
	got := dataToBytes("4E")
	if !bytes.Equal(got, want) {
		t.Errorf("got %#x, want %#x", got, want)
	}

	want = []byte{
		0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
		0x05, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
	}
	got = dataToBytes("1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1 5 1 1 1 1 1 1 1 1 1 1 1 1 1 1 1")
	if !bytes.Equal(got, want) {
		t.Errorf("got %#x, want %#x", got, want)
	}
}
