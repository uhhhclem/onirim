package onirim

import (
	"github.com/uhhhclem/mse/src/interact"
)

type Game struct {
	*interact.Game
	Deck Deck
	Hand Deck
	Row  Deck
}

func NewGame() (*Game, error) {
	g := &Game{
		Game: interact.NewGame(),
	}
	g.Deck = MakeDeck()
	g.Hand = make(Deck, 0, 5)
	if err := FillHand(&g.Deck, &g.Hand); err != nil {
		return nil, err
	}
	g.Row = make(Deck, 0, len(g.Deck))
	return g, nil
}

const (
	startOfTurn interact.GameState = "startOfTurn"
)
