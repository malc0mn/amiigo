package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/malc0mn/amiigo/apii"
	"sort"
	"strings"
)

// formatAmiiboInfo formats part of an apii.AmiiboInfo struct for display in a box.
func formatAmiiboInfo(ai *apii.AmiiboInfo) []byte {
	releases := make([]string, len(ai.Release))
	i := 0
	for c, d := range ai.Release {
		releases[i] = "  " + strings.ToUpper(c) + ": " + d
		i++
	}
	sort.Strings(releases)

	pref := "\n  "
	// Maps are unordered, hence this approach.
	info := []string{
		"ID:", pref + "0x" + ai.Head + ai.Tail,
		"Character:", pref + ai.Character,
		"Name:", pref + ai.Name,
		"Type:", pref + ai.Type,
		"Amiibo Series:", pref + ai.AmiiboSeries,
		"Game series:", pref + ai.GameSeries,
		"Release dates:", "\n" + strings.Join(releases, "\n"),
	}

	return encodeWithLabelToBytes(info)
}

// formatAmiiboUsage formats the usage info of all games from an apii.AmiiboInfo struct for display
// in a box.
func formatAmiiboUsage(ai []apii.AmiiboInfo) []byte {
	var usage []string
	for _, cu := range ai {
		if cu.GamesSwitch != nil {
			usage = append(usage, "Switch:", "\n")
			for _, i := range cu.GamesSwitch {
				usage = append(usage, formatGameInfo(&i)...)
			}
		}
		if cu.GamesWiiU != nil {
			usage = append(usage, "WiiU:", "\n")
			for _, i := range cu.GamesWiiU {
				usage = append(usage, formatGameInfo(&i)...)
			}
		}
		if cu.Games3DS != nil {
			usage = append(usage, "3DS:", "\n")
			for _, i := range cu.Games3DS {
				usage = append(usage, formatGameInfo(&i)...)
			}
		}
	}

	return encodeWithLabelToBytes(usage)
}

// formatGameInfo formats an apii.GameInfo struct to a sting array.
func formatGameInfo(gi *apii.GameInfo) []string {
	usage := []string{"  Game:  ", gi.GameName}
	split := " / "
	for _, u := range gi.AmiiboUsage {
		listItem := ""
		if strings.Count(u.Usage, split) > 0 {
			listItem = "- "
		}
		usage = append(usage, "  Usage: ", listItem+strings.Replace(u.Usage, split, "\n         "+listItem, -1), "  Write: ", fmt.Sprintf("%v", u.Write), "", "\n")
	}

	return usage
}

// encodeWithLabelToBytes encodes a string array composed of labels and values to a byte array for display
// in a box.
func encodeWithLabelToBytes(s []string) []byte {
	var res []byte
	for i, v := range s {
		if i%2 == 0 {
			res = append(res, encodeStringCellWithAttrs(v, tcell.AttrBold|tcell.AttrDim, "")...)
			continue
		}
		res = append(res, encodeStringCell(v)...)
	}

	return res
}
