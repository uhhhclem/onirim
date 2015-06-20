package onirim

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/uhhhclem/mse/src/interact"
)

type Game struct {
	*interact.Game
	Done    bool
	Deck    Deck
	Hand    Deck
	Row     Deck
	Discard Deck
	Limbo   Deck
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
	g.Discard = make(Deck, 0, len(g.Deck))
	g.Limbo = make(Deck, 0, len(g.Deck))
	return g, nil
}

func (g *Game) RunLocal() {
	g.State = startOfTurn

	go func() {
		for {
			s := <-g.NextStatus
			if s == nil {
				return
			}
			fmt.Println(s.Message)
		}
	}()

	go func() {
		for {
			p := <-g.NextPrompt
			if p == nil {
				return
			}
			fmt.Println(p.Message)
			for _, c := range p.Choices {
				fmt.Printf("  %s: %s\n", c.Key, c.Name)
			}
		GetChoice:
			for {
				var key string
				n, err := fmt.Scanf("%s\n", &key)
				if err != nil || n != 1 {
					fmt.Println(n, err)
					continue
				}
				key = strings.ToUpper(key)
				for _, c := range p.Choices {
					if key == c.Key {
						g.NextChoice <- c
						break GetChoice
					}
				}
			}
		}
	}()

	for {
		if g.Done {
			break
		}
		h := handlers[g.State]
		g.State = h(g)
	}
}

const (
	startOfTurn   interact.GameState = "startOfTurn"
	playOrDiscard                    = "playOrDiscard"
	endOfTurn                        = "endOfTurn"
	endOfGame                        = "endOfGame"
	keyDrawn                         = "keyDrawn"
	dreamDrawn                       = "dreamDrawn"
)

type handler func(g *Game) interact.GameState

var handlers = map[interact.GameState]handler{
	startOfTurn:   handleStartOfTurn,
	playOrDiscard: handlePlayOrDiscard,
	endOfTurn:     handleEndOfTurn,
	endOfGame:     handleEndOfGame,
	keyDrawn:      handleKeyDrawn,
	dreamDrawn:    handleDreamDrawn,
}

func handleStartOfTurn(g *Game) interact.GameState {
	log.Println(startOfTurn)
	playable := make(Deck, 0, 5)
	for _, c := range g.Hand {
		if g.isPlayable(c) {
			playable = append(playable, c)
		}
	}
	g.NewPrompt("Select card to play or discard")
	for i, c := range playable {
		g.AddChoice(
			fmt.Sprintf("P%d", i+1),
			fmt.Sprintf("Play %s", c))
	}
	for i, c := range g.Hand {
		g.AddChoice(
			fmt.Sprintf("D%d", i+1),
			fmt.Sprintf("Discard %s", c))
	}
	g.SendPrompt()
	return playOrDiscard
}

func handlePlayOrDiscard(g *Game) interact.GameState {
	log.Printf("handlePlayOrDiscard")
	c := <-g.NextChoice
	action := string(c.Key[0])
	index := string(c.Key[1])
	i, err := strconv.Atoi(index)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	i = i - 1
	card := g.Hand.RemoveCardAt(i)
	if action == "P" {
		g.Row.AddCard(card)
	} else {
		g.Discard.AddCard(card)
	}
	return endOfTurn
}

func handleEndOfTurn(g *Game) interact.GameState {
	log.Printf("handleEndOfTurn")
	log.Printf("\nHand    : %s\nRow     : %s\nDiscard : %s\n", g.Hand, g.Row, g.Discard)
	if len(g.Hand) == 5 {
		if len(g.Limbo) > 0 {
			g.Deck = append(g.Deck, g.Limbo...)
			g.Deck.Shuffle()
		}
		return startOfTurn
	}
	c, err := g.Deck.DrawCard()
	if err != nil {
		log.Println(err)
		return endOfGame
	}
	g.Logf("Drew %s", c)
	switch c.Class {
	case Labyrinth:
		if c.Symbol == Key {
			return keyDrawn
		}
		g.Hand.AddCard(c)
		return endOfTurn
	case Door:
		g.Limbo.AddCard(c)
		return endOfTurn
	default:
		return dreamDrawn
	}
}

func handleKeyDrawn(g *Game) interact.GameState {
	log.Print(keyDrawn)
	return endOfGame
}

func handleDreamDrawn(g *Game) interact.GameState {
	log.Print(dreamDrawn)
	return endOfGame
}

func handleEndOfGame(g *Game) interact.GameState {
	log.Printf(endOfGame)
	g.Done = true
	return endOfGame
}

func (g *Game) isPlayable(c *Card) bool {
	if len(g.Row) == 0 {
		return true
	}
	if c.Class != Labyrinth {
		return false
	}
	return c.Symbol != g.Row.LastCard().Symbol
}
