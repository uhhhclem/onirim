package onirim

import (
	"bytes"
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

func Print(s ...interface{}) {
	printChan <- fmt.Sprint(s)
}

func Println(s ...interface{}) {
	printChan <- fmt.Sprintln(s)
}

func Printf(f string, data ...interface{}) {
	printChan <- fmt.Sprintf(f, data...)
}

type Game struct {
	*interact.Game
	Done      bool           // Is the game over?
	Deck      Deck           // Your deck.
	Hand      Deck           // Your hand.  Ordering is insignificant.
	Row       Deck           // The row of cards that you play to.
	Discard   Deck           // The discard pile.
	Limbo     Deck           // Where Doors and Nightmares wait for reshuffling.
	Doors     Deck           // Discovered doors.  Ordering is insignificant.
	Drawn     *Card          // The last card drawn.
	FoundDoor map[*Card]bool // Labyrinth cards used to discover a Door.
}

func NewGame() (*Game, error) {
	g := &Game{
		Game:      interact.NewGame(),
		Deck:      MakeDeck(),
		FoundDoor: make(map[*Card]bool),
	}
	if err := g.fillHand(); err != nil {
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
			buf := &bytes.Buffer{}
			fmt.Fprintln(buf, p.Message)
			for _, c := range p.Choices {
				fmt.Fprintf(buf, "  %s: %s\n", c.Key, c.Name)
			}
			Print(buf.String())
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
	c := <-g.NextChoice
	action, i := parseKey(c.Key)
	card := g.Hand.RemoveCardAt(i)

	if action == "D" {
		g.discard(card)
		if card.Symbol == Key {
			return prophecy
		}
		return endOfTurn
	}

	g.playCard(card)
	if g.isDoorDiscovered() && g.playDoor(card.Color) {
		for i = 0; i < 3; i++ {
			index := len(g.Row) - i - 1
			g.FoundDoor[g.Row[index]] = true
		}
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
	// Unlike g.fillHand(), this draws exactly 5 cards under all circumstances.
	for i := 0; i < 5; i++ {
		c, err := g.drawCard()
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
	g.discard(c)

	for len(temp) > 1 {
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
		g.placeOnDeck(c)
	}
	g.placeOnDeck(temp[0])
	return endOfTurn
}

// parseKey parses a key like "P2" into an action of "P" and an index of 1.
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

func (g *Game) isDoorDiscovered() bool {
	if len(g.Row) == 0 {
		return false
	}
	color := g.Row[len(g.Row)-1].Color
	var found Deck

	for i := len(g.Row) - 1; i >= 0 && len(found) < 3; i-- {
		c := g.Row[i]
		if g.FoundDoor[c] || c.Color != color {
			return false
		}
		found = append(found, c)
	}
	if len(found) != 3 {
		return false
	}
	return true
}

// playDoor returns true if it can find and play a Door of the given color.
func (g *Game) playDoor(color ColorEnum) bool {
	for i, c := range g.Deck {
		if c.Class != Door || c.Color != color {
			continue
		}
		g.Deck.RemoveCardAt(i)
		g.Doors.AddCard(c)
		g.Logf("Played %s door", color)
		return true
	}
	return false
}

func handleEndOfTurn(g *Game) interact.GameState {
	Printf(`

Hand    : %s
Row     : %s
Discard : %s
Doors   : %s

`, g.Hand, g.Row, g.Discard, g.Doors)

	if len(g.Hand) == 5 {
		g.shuffleLimboIntoDeck()
		return startOfTurn
	}

	var err error
	g.Drawn, err = g.drawCard()
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
	color := g.Drawn.Color
	index, keyCard := g.matchingKeyInHand(color)
	if keyCard == nil {
		g.moveToLimbo(g.Drawn)
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
		g.discard(keyCard)
		g.Hand.RemoveCardAt(index)
	} else {
		g.discard(g.Drawn)
	}
	g.Drawn = nil
	return endOfTurn
}

func (g *Game) matchingKeyInHand(color ColorEnum) (int, *Card) {
	for index, card := range g.Hand {
		if card.Symbol == Key && card.Color == color {
			return index, card
		}
	}
	return 0, nil
}

func handleDreamDrawn(g *Game) interact.GameState {
	g.NewPrompt("You've drawn a Nightmare:")
	for i, card := range g.Hand {
		if card.Symbol == Key {
			g.AddChoice(fmt.Sprintf("K%d", i+1), fmt.Sprintf("Discard %s from hand", card))
		}
	}
	for i, card := range g.Doors {
		g.AddChoice(fmt.Sprintf("R%d", i+1), fmt.Sprintf("Move %s from to Limbo", card))
	}
	g.AddChoice("H0", "Discard your hand")
	g.AddChoice("T0", "Discard cards from the deck")
	g.SendPrompt()
	choice := <-g.NextChoice
	action, index := parseKey(choice.Key)
	switch action {
	case "K":
		c := g.Hand.RemoveCardAt(index)
		g.discard(c)
	case "R":
		c := g.Doors.RemoveCardAt(index)
		g.moveToLimbo(c)
	case "H":
		for _, c := range g.Hand {
			g.discard(c)
		}
		if err := g.fillHand(); err != nil {
			return endOfGame
		}
	case "T":
		for i := 0; i < 5; i++ {
			c, err := g.drawCard()
			if err != nil {
				return endOfGame
			}
			if c.Class == Labyrinth {
				g.discard(c)
			} else {
				g.moveToLimbo(c)
			}
		}
	}
	return endOfTurn
}

func (g *Game) shuffleLimboIntoDeck() {
	if len(g.Limbo) == 0 {
		return
	}
	g.Deck = append(g.Deck, g.Limbo...)
	g.Limbo = nil
	g.Deck.Shuffle()
}

func handleEndOfGame(g *Game) interact.GameState {
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

func (g *Game) drawCard() (*Card, error) {
	c, err := g.Deck.DrawCard()
	if err == nil {
		g.Logf("Drew %s", c)
	}
	return c, err
}

func (g *Game) discard(c *Card) {
	g.Discard.AddCard(c)
	g.Logf("Discarded %s", c)
}

func (g *Game) placeOnDeck(c *Card) {
	g.Deck = append(Deck{c}, g.Deck...)
	g.Logf("Placed %s on deck", c)
}

func (g *Game) playCard(c *Card) {
	g.Row.AddCard(c)
	g.Logf("Played %s to row", c)
}

func (g *Game) moveToLimbo(c *Card) {
	g.Limbo.AddCard(c)
	g.Logf("Moved %s to Limbo", c)
}

func (g *Game) fillHand() error {
	g.Hand = nil
	for i := 0; i < 5; {
		c, err := g.drawCard()
		if err != nil {
			return err
		}
		if c.Class != Labyrinth {
			g.moveToLimbo(c)
			continue
		}
		g.Hand.AddCard(c)
		i++
	}
	return nil
}
