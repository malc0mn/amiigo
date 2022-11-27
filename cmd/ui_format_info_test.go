package main

import (
	"bytes"
	"github.com/gdamore/tcell/v2"
	"github.com/malc0mn/amiigo/apii"
	"testing"
)

func TestFormatAmiiboUsage(t *testing.T) {
	// TODO: find better way to render and test, templates maybe?
	/*ai := []apii.AmiiboInfo{
		{
			Games3DS:    []apii.GameInfo{{GameName: "test 3ds", AmiiboUsage: []apii.Usage{{Usage: "do stuff"}}}},
			GamesSwitch: []apii.GameInfo{{GameName: "test switch", AmiiboUsage: []apii.Usage{{Usage: "do stuff / and lots more"}}}},
			GamesWiiU:   []apii.GameInfo{{GameName: "test wiiu", AmiiboUsage: []apii.Usage{{Usage: "do stuff", Write: true}}}},
		},
	}

	got := formatAmiiboUsage(ai)

	want := encodeStringCellWithAttrs("Switch:", tcell.AttrBold|tcell.AttrDim, "\n")
	want = append(want, encodeWithLabelToBytes([]string{"  Game:  ", "test switch", "  Usage: ", "- do stuff\n         - and lots more", "  Write: ", "false", "", "\n"})...)
	want = append(want, encodeStringCellWithAttrs("WiiU:", tcell.AttrBold|tcell.AttrDim, "\n")...)
	want = append(want, encodeWithLabelToBytes([]string{"  Game:  ", "test wiiu", "  Usage: ", "do stuff", "  Write: ", "true", "", "\n"})...)
	want = append(want, encodeStringCellWithAttrs("3DS:", tcell.AttrBold|tcell.AttrDim, "\n")...)
	want = append(want, encodeWithLabelToBytes([]string{"  Game:  ", "test 3ds", "  Usage: ", "do stuff", "  Write: ", "false", "", "\n"})...)

	if len(got) != len(want) {
		t.Fatalf("sizes differ:\ngot\n%s\nwant\n%s", got, want)
	}

	if !bytes.Equal(got, want) {
		t.Errorf("got\n%v\nwant\n%v", got, want)
	}*/
}

func TestFormatGameInfo(t *testing.T) {
	tests := []struct {
		gi   *apii.GameInfo
		want []string
	}{
		{
			gi:   &apii.GameInfo{GameName: "test", AmiiboUsage: []apii.Usage{{Usage: "do stuff"}}},
			want: []string{"  Game:  ", "test", "  Usage: ", "do stuff", "  Write: ", "false", "", "\n"},
		},
		{
			gi:   &apii.GameInfo{GameName: "test", AmiiboUsage: []apii.Usage{{Usage: "do stuff / and lots more"}}},
			want: []string{"  Game:  ", "test", "  Usage: ", "- do stuff\n         - and lots more", "  Write: ", "false", "", "\n"},
		},
	}

	for _, test := range tests {
		got := formatGameInfo(test.gi)

		if len(got) != len(test.want) {
			t.Fatalf("sizes differ:\n%s\n%s", got, test.want)
		}

		for i, g := range got {
			if g != test.want[i] {
				t.Errorf("got '%s', want '%s'", got, test.want)
			}
		}
	}
}

func TestEncodeWithLabelToBytes(t *testing.T) {
	got := encodeWithLabelToBytes([]string{"Hello: ", "world"})

	want := encodeStringCellWithAttrs("Hello: ", tcell.AttrBold|tcell.AttrDim, "")
	want = append(want, encodeStringCell("world")...)

	if !bytes.Equal(got, want) {
		t.Errorf("got\n%v\nwant\n%v", got, want)
	}
}
