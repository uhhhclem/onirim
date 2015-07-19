package onirim

import (
	"fmt"
)

type Board struct {
	Hand           []string
	Discard        []string
	Doors          []string
	Row            []string
	CardsRemaining int
	Done           bool
	Won            bool
}

var classKey = map[ClassEnum]string{
	Dream:     "D",
	Door:      "R",
	Labyrinth: "L",
}

var colorKey = map[ColorEnum]string{
	Red:   "R",
	Blue:  "B",
	Green: "G",
	Brown: "Y",
}

var symbolKey = map[SymbolEnum]string{
	Key:  "K",
	Sun:  "S",
	Moon: "M",
}

func (c *Card) key() string {
	if c.Class == Dream {
		return "DN"
	}
	return fmt.Sprintf("%s%s%s", classKey[c.Class], colorKey[c.Color], symbolKey[c.Symbol])
}

func (g *Game) GetBoard() Board {
	b := Board{
		Hand:           make([]string, len(g.Hand)),
		Discard:        make([]string, len(g.Discard)),
		Doors:          make([]string, len(g.Doors)),
		Row:            make([]string, len(g.Row)),
		CardsRemaining: len(g.Deck),
		Done:           g.Done,
		Won:            g.Won,
	}
	for i, c := range g.Hand {
		b.Hand[i] = c.key()
	}
	for i, c := range g.Discard {
		b.Discard[i] = c.key()
	}
	for i, c := range g.Doors {
		b.Doors[i] = c.key()
	}
	for i, c := range g.Row {
		b.Row[i] = c.key()
	}
	return b
}
