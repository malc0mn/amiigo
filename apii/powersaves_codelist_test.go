package apii

import (
	"bytes"
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

func TestNewCheatList(t *testing.T) {
	file := "codelist.xml"
	c, err := NewCheatList(readFile(t, file))
	if err != nil {
		t.Errorf("could not unmarshal file %s, error %s", file, err)
	}

	wanti := 9
	goti := len(c.Games)
	if goti != wanti {
		t.Errorf("len(c.Games) got %d, want %d", goti, wanti)
	}

	wants := "Shovel Knight (3DS)"
	gots := c.Games[5].Name
	if gots != wants {
		t.Errorf("c.Games.Name got %s, want %s", gots, wants)
	}

	wanti = 8
	goti = c.Games[5].Idx
	if goti != wanti {
		t.Errorf("c.Games.Id got %d, want %d", goti, wanti)
	}

	wanti = 1
	goti = len(c.Games[5].Folder.Cheats)
	if goti != wanti {
		t.Errorf("len(c.Games.Folders.Cheats) got %d, want %d", goti, wanti)
	}

	wantsl := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 98, 99}
	gotsl := c.GameIds
	if !equal(gotsl, wantsl) {
		t.Errorf("c.GameIds got %v, want %v", gotsl, wantsl)
	}

	wanti = 699
	goti = len(c.Toys)
	if goti != wanti {
		t.Errorf("len(c.Toys) got %d, want %d", goti, wanti)
	}

	wants = "MSSS - Rosalina (Golf)"
	gots = c.Toys[186].Name
	if gots != wants {
		t.Errorf("c.Toys.Name got %s, want %s", gots, wants)
	}

	wanti = 0x09cf040102b70e02
	goti = int(c.Toys[186].Id)
	if goti != wanti {
		t.Errorf("c.Toys.Id got %#016x, want %#016x", goti, wanti)
	}

	wanti = 4
	goti = len(c.Lists)
	if goti != wanti {
		t.Errorf("len(c.Lists) got %d, want %d", goti, wanti)
	}

	wanti = 2
	goti = c.Lists[2].Idx
	if goti != wanti {
		t.Errorf("c.Lists.Id got %d, want %d", goti, wanti)
	}

	wanti = 34
	goti = len(c.Lists[2].Cheats)
	if goti != wanti {
		t.Errorf("c.Lists.Cheats got %d, want %d", goti, wanti)
	}

	wants = "Agility (Can be used in Four slots Max)"
	gots = c.Lists[3].Cheats[5].Name
	if gots != wants {
		t.Errorf("c.Lists.Cheats.Name got %s, want %s", gots, wants)
	}

	wants = ""
	gots = c.Lists[3].Cheats[5].Desc
	if gots != wants {
		t.Errorf("c.Lists.Cheats.Desc got %s, want %s", gots, wants)
	}

	wanti = 0
	goti = c.Lists[3].Cheats[5].Idx
	if goti != wanti {
		t.Errorf("c.Lists.Cheats.Id got %d, want %d", goti, wanti)
	}

	wants = "6"
	gots = c.Lists[3].Cheats[5].Data
	if gots != wants {
		t.Errorf("c.Lists.Cheats.Data got %s, want %s", gots, wants)
	}

	wanti = 26
	goti = len(c.Folders)
	if goti != wanti {
		t.Errorf("len(c.Folders) got %d, want %d", goti, wanti)
	}

	wants = "Special Moves"
	gots = c.Folders[4].Name
	if gots != wants {
		t.Errorf("c.Folders.Name got %s, want %s", gots, wants)
	}

	wantsl = []int{0, 2}
	gotsl = c.Folders[4].GameIds
	if !equal(gotsl, wantsl) {
		t.Errorf("c.Folders.GameIds got %v, want %v", gotsl, wantsl)
	}

	wants = "f004"
	gots = c.Folders[4].Idx
	if gots != wants {
		t.Errorf("c.Folders.Id got %s, want %s", gots, wants)
	}

	wants = "Slider"
	gots = c.Folders[4].Type
	if gots != wants {
		t.Errorf("c.Folders.Type got %s, want %s", gots, wants)
	}

	wanti = 4
	goti = len(c.Folders[4].Cheats)
	if goti != wanti {
		t.Errorf("c.Folders.Folders got %d, want %d", goti, wanti)
	}

	wants = "Up"
	gots = c.Folders[4].Cheats[2].Name
	if gots != wants {
		t.Errorf("c.Folders.Folders.Name got %s, want %s", gots, wants)
	}

	wantsl = []int{0, 2}
	gotsl = c.Folders[4].Cheats[2].GamesIds
	if !equal(gotsl, wantsl) {
		t.Errorf("c.Folders.Folders.GameIds got %v, want %v", gotsl, wantsl)
	}

	wanti = 0
	goti = c.Folders[4].Cheats[2].Idx
	if goti != wanti {
		t.Errorf("c.Folders.Folders.Id got %d, want %d", goti, wanti)
	}

	wanti = 1
	goti = len(c.Folders[4].Cheats[2].Codes)
	if goti != wanti {
		t.Errorf("c.Folders.Folders.Codes got %d, want %d", goti, wanti)
	}

	wants = "110e7"
	gots = c.Folders[4].Cheats[2].Codes[0].Addr
	if gots != wants {
		t.Errorf("c.Folders.Folders.Codes.Addr got %s, want %s", gots, wants)
	}

	wants = "1 3"
	gots = c.Folders[4].Cheats[2].Codes[0].Data
	if gots != wants {
		t.Errorf("c.Folders.Folders.Codes.Data got %s, want %s", gots, wants)
	}

	wanti = 3
	goti = c.Folders[14].ListIdx
	if goti != wanti {
		t.Errorf("c.Folders.ListIdx got %d, want %d", goti, wanti)
	}
	wants = "210e1"
	gots = c.Folders[14].Addr
	if gots != wants {
		t.Errorf("c.Folders.Addr got %s, want %s", gots, wants)
	}

	wanti = c.Folders[14].ListIdx
	goti = c.Folders[14].List.Idx
	if goti != wanti {
		t.Errorf("c.Folders.List got %d, want %d", goti, wanti)
	}

	gotp := c.Folders[4].List
	if gotp != nil {
		t.Errorf("c.Folders.List got %v, want nil", gotp)
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

func TestInt32ToByte(t *testing.T) {
	want := []byte{0xff, 0xff, 0xff, 0x3a}
	got := int32ToByte(-198)
	if !bytes.Equal(got, want) {
		t.Errorf("got %#x, want %#x", got, want)
	}
}
