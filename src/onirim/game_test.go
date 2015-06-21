package onirim

import (
	"testing"
)

func TestisDoorDiscovered(t *testing.T) {
	var tcs = []struct {
		colors []ColorEnum
		found  []bool
		result bool
		desc   string
	}{
		{
			colors: []ColorEnum{},
			found:  []bool{},
			result: false,
			desc:   "no cards in row",
		},
		{
			colors: []ColorEnum{Red, Red},
			found:  []bool{false, false},
			result: false,
			desc:   "two cards in row",
		},
		{
			colors: []ColorEnum{Blue, Red, Red},
			found:  []bool{false, false, false},
			result: false,
			desc:   "not all colors match",
		},
		{
			colors: []ColorEnum{Green, Green, Green},
			found:  []bool{true, false, false},
			result: false,
			desc:   "one card was previously used",
		},
		{
			colors: []ColorEnum{Brown, Green, Green, Green},
			found:  []bool{false, false, false, false},
			result: true,
			desc:   "happy case",
		},
	}

	for _, tc := range tcs {
		if len(tc.colors) != len(tc.found) {
			t.Fatal("Improper test case.")
		}
		g := &Game{FoundDoor: make(map[*Card]bool)}
		for i, color := range tc.colors {
			card := &Card{Color: color}
			g.Row.AddCard(card)
			g.FoundDoor[card] = tc.found[i]
		}
		if got, want := g.isDoorDiscovered(), tc.result; got != want {
			t.Errorf("%s: got %t, want %t", tc.desc, got, want)
		}
	}
}
