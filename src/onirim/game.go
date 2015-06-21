package onirim

import (
	"fmt"
	"os"
	"strconv"

	"github.com/uhhhclem/mse/src/interact"
)

var printChan chan string

func init() {
	printChan = make(chan string)
	go func() {
		for {
			s := <-printChan
			fmt.Print(s)
		}
	}()
}

func Println(s ...interface{}) {
	printChan <- fmt.Sprintln(s)
}

func Printf(f string, data ...interface{}) {
	printChan <- fmt.Sprintf(f, data...)
}

type Game struct {
	*interact.Game
	Done    bool
	Deck    Deck
	Hand    Deck
	Row     Deck
	Discard Deck
	Limbo   Deck
	Doors   Deck
	Drawn   *Card
}

func NewGame() (*Game, error) {
	g := &Game{
		Game: interact.NewGame(),
	}
	g.Deck = MakeDeck()
	if err := FillHand(&g.Deck, &g.Hand); err != nil {
		return nil, err
	}
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
			Println(s.Message)
		}
	}()

	go func() {
		for {
			p := <-g.NextPrompt
			if p == nil {
				return
			}
			Println(p.Message)
			for _, c := range p.Choices {
				Printf("  %s: %s\n", c.Key, c.Name)
			}
			for {
				var key string
				n, err := fmt.Scanf("%s\n", &key)
				if err != nil || n != 1 {
					Println(n, err)
					continue
				}
				if err := g.MakeChoice(key); err != nil {
					Println(err)
					continue
				}
				break
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
	doorDrawn                        = "doorDrawn"
	dreamDrawn                       = "dreamDrawn"
	prophecy                         = "prophecy"
)

type handler func(g *Game) interact.GameState

var handlers = map[interact.GameState]handler{
	startOfTurn:   handleStartOfTurn,
	playOrDiscard: handlePlayOrDiscard,
	endOfTurn:     handleEndOfTurn,
	endOfGame:     handleEndOfGame,
	doorDrawn:     handleDoorDrawn,
	dreamDrawn:    handleDreamDrawn,
	prophecy:      handleProphecy,
}

func handleStartOfTurn(g *Game) interact.GameState {
	Println(startOfTurn)
	playable := make(map[*Card]bool)
	for _, c := range g.Hand {
		playable[c] = g.isPlayable(c)
	}
	g.NewPrompt("Select card to play or discard")
	for i, c := range g.Hand {
		if playable[c] {
			g.AddChoice(
				fmt.Sprintf("P%d", i+1),
				fmt.Sprintf("Play %s", c))
		}
	}
	for i, c := range g.Hand {
		f := "Discard %s"
		if c.Class == Labyrinth && c.Symbol == Key {
			f = "Discard %s and trigger prophecy"
		}
		g.AddChoice(
			fmt.Sprintf("D%d", i+1),
			fmt.Sprintf(f, c))
	}
	g.SendPrompt()
	return playOrDiscard
}

func handlePlayOrDiscard(g *Game) interact.GameState {
	Printf(playOrDiscard)
	c := <-g.NextChoice
	action, i := parseKey(c.Key)
	card := g.Hand.RemoveCardAt(i)

	if action == "D" {
		g.Discard.AddCard(card)
		if card.Symbol == Key {
			return prophecy
		}
		return endOfTurn
	}

	g.Row.AddCard(card)
	if g.IsDoorDiscovered() {
		g.PlayDoor(card.Color)
		if len(g.Doors) == 8 {
			g.Done = true
			return endOfGame
		}
	}
	return endOfTurn
}

func handleProphecy(g *Game) interact.GameState {
	var temp Deck
	g.Logf("Prophecy triggered")
	for i := 0; i < 5; i++ {
		c, err := g.Deck.DrawCard()
		if err != nil {
			return endOfGame
		}
		temp.AddCard(c)
	}
	g.NewPrompt("Select one card to discard:")
	for i, c := range temp {
		g.AddChoice(
			fmt.Sprintf("D%d", i+1),
			fmt.Sprintf("Discard %s", c))
	}
	g.SendPrompt()
	choice := <-g.NextChoice
	_, index := parseKey(choice.Key)
	c := temp.RemoveCardAt(index)
	g.Discard.AddCard(c)

	for len(temp) > 0 {
		g.NewPrompt("Select card to place on top of deck:")
		for i, c := range temp {
			g.AddChoice(
				fmt.Sprintf("P%d", i+1),
				fmt.Sprintf("Place %s on deck", c))
		}
		g.SendPrompt()
		choice := <-g.NextChoice
		_, index := parseKey(choice.Key)
		c := temp.RemoveCardAt(index)
		g.Deck = append(Deck{c}, g.Deck...)
	}
	return endOfTurn
}

func parseKey(key string) (string, int) {
	action := string(key[0])
	index := string(key[1])
	i, err := strconv.Atoi(index)
	if err != nil {
		Println(err)
		os.Exit(1)
	}
	return action, i - 1
}

func (g *Game) IsDoorDiscovered() bool {
	last := len(g.Row) - 1
	if last < 2 {
		return false
	}
	color := g.Row[last].Color
	return color == g.Row[last-1].Color && color == g.Row[last-1].Color
}

func (g *Game) PlayDoor(color ColorEnum) {
	for i, c := range g.Deck {
		if c.Class != Door || c.Color != color {
			continue
		}
		g.Deck.RemoveCardAt(i)
		g.Doors.AddCard(c)
		g.Logf("Played a %s door.", c.Color)
		return
	}
}

func handleEndOfTurn(g *Game) interact.GameState {
	Printf(endOfTurn)
	Printf(`

Hand    : %s
Row     : %s
Discard : %s
Doors   : %s

`, g.Hand, g.Row, g.Discard, g.Doors)

	if len(g.Hand) == 5 {
		if len(g.Limbo) > 0 {
			g.Deck = append(g.Deck, g.Limbo...)
			g.Deck.Shuffle()
		}
		return startOfTurn
	}

	var err error
	g.Drawn, err = g.Deck.DrawCard()
	if err != nil {
		Println(err)
		return endOfGame
	}
	g.Logf("Drew %s", g.Drawn)
	switch g.Drawn.Class {
	case Labyrinth:
		g.Hand.AddCard(g.Drawn)
		g.Drawn = nil
		return endOfTurn
	case Door:
		return doorDrawn
	case Dream:
		return dreamDrawn
	default:
		fmt.Printf("\nUnknown class: %s", g.Drawn.Class)
		return endOfGame
	}
}

func handleDoorDrawn(g *Game) interact.GameState {
	Println(doorDrawn)
	color := g.Drawn.Color
	index, keyCard := g.MatchingKeyInHand(color)
	if keyCard == nil {
		g.Limbo.AddCard(g.Drawn)
		g.Drawn = nil
		return endOfTurn
	}
	g.NewPrompt("You've drawn a door:")
	g.AddChoice("Y", fmt.Sprintf("Discard %s Key to play %s Door", color, color))
	g.AddChoice("N", fmt.Sprintf("Keep %s Key and move %s Door to Limbo", color, color))
	g.SendPrompt()
	choice := <-g.NextChoice

	if choice.Key == "Y" {
		g.Doors.AddCard(g.Drawn)
		g.Discard.AddCard(keyCard)
		g.Hand.RemoveCardAt(index)
	} else {
		g.Discard.AddCard(g.Drawn)
	}
	g.Drawn = nil
	return endOfTurn
}

func (g *Game) MatchingKeyInHand(color ColorEnum) (int, *Card) {
	for index, card := range g.Hand {
		if card.Symbol == Key && card.Color == color {
			return index, card
		}
	}
	return 0, nil
}

func handleDreamDrawn(g *Game) interact.GameState {
	Println(dreamDrawn)
	g.NewPrompt("You've drawn a Nightmare:")
	for i, card := range g.Hand {
		if card.Symbol == Key {
			g.AddChoice(fmt.Sprintf("K%d", i+1), fmt.Sprintf("Discard %s from hand", card))
		}
	}
	for i, card := range g.Doors {
		g.AddChoice(fmt.Sprintf("R%d", i+1), fmt.Sprintf("Discard %s from discovered Doors", card))
	}
	g.AddChoice("H0", "Discard your hand")
	g.AddChoice("T0", "Discard cards from the deck")
	g.SendPrompt()
	choice := <-g.NextChoice
	action, index := parseKey(choice.Key)
	switch action {
	case "K":
		c := g.Hand.RemoveCardAt(index)
		g.Discard.AddCard(c)
	case "R":
		c := g.Doors.RemoveCardAt(index)
		g.Discard.AddCard(c)
	case "H":
		for _, c := range g.Hand {
			g.Discard.AddCard(c)
		}
		g.Hand = nil
	case "T":
		for i := 0; i < 5; i++ {
			c, err := g.Deck.DrawCard()
			if err != nil {
				return endOfGame
			}
			if c.Class == Labyrinth {
				g.Discard.AddCard(c)
			} else {
				g.Limbo.AddCard(c)
			}
		}
		g.ShuffleLimboIntoDeck()
	}
	return endOfTurn
}

func (g *Game) ShuffleLimboIntoDeck() {
	if len(g.Limbo) == 0 {
		return
	}
	g.Deck = append(g.Deck, g.Limbo...)
	g.Deck.Shuffle()
}

func handleEndOfGame(g *Game) interact.GameState {
	Printf(endOfGame)
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
