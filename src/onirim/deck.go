// Package onirim contains logic for playing Shadi Torbey's game Onirim.
package onirim

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// ColorEnum represents the color of a card.
type ColorEnum int

const (
	NoColor ColorEnum = iota
	Red
	Blue
	Green
	Brown
)

// Colors contains the names of the colors.
var colors = map[ColorEnum]string{
	NoColor: "None",
	Red:     "Red",
	Blue:    "Blue",
	Green:   "Green",
	Brown:   "Brown",
}

// String returns the name of the color.
func (c ColorEnum) String() string {
	return colors[c]
}

// Class represents the class of a card.
type ClassEnum int

const (
	Dream ClassEnum = iota
	Door
	Labyrinth
)

var classes = map[ClassEnum]string{
	Dream:     "Dream",
	Door:      "Door",
	Labyrinth: "Labyrinth",
}

func (c ClassEnum) String() string {
	return classes[c]
}

type SymbolEnum int

const (
	NoSymbol SymbolEnum = iota
	Key
	Sun
	Moon
)

var symbols = map[SymbolEnum]string{
	NoSymbol: "None",
	Key:      "Key",
	Sun:      "Sun",
	Moon:     "Moon",
}

func (s SymbolEnum) String() string {
	return symbols[s]
}

// Card represents an Onirim card.
type Card struct {
	Class  ClassEnum
	Color  ColorEnum
	Symbol SymbolEnum
}

func (c *Card) String() string {
	switch c.Class {
	case Dream:
		return "D:NM"
	case Labyrinth:
		return fmt.Sprintf("L:%s%s", c.Color.String()[0:1], c.Symbol.String()[0:1])
	case Door:
		return fmt.Sprintf("R:%s", c.Color.String()[0:1])
	default:
		return ""
	}
}

// Deck represents any ordered collection of cards:  the deck, a hand, a pile, etc.
type Deck []*Card

func (d Deck) String() string {
	s := make([]string, 0, len(d))
	for i, c := range d {
		s = append(s, fmt.Sprintf("%d-%s", i+1, c))
	}
	return strings.Join(s, ", ")
}

func addLabyrinthCards(d *Deck, c ColorEnum, sunCnt int) {
	for i := 0; i < 3; i++ {
		d.AddCard(&Card{Labyrinth, c, Key})
	}
	for i := 0; i < 4; i++ {
		d.AddCard(&Card{Labyrinth, c, Moon})
	}
	for i := 0; i < sunCnt; i++ {
		d.AddCard(&Card{Labyrinth, c, Sun})
	}
}

func addDoorCards(d *Deck, c ColorEnum) {
	d.AddCard(&Card{Class: Door, Color: c})
	d.AddCard(&Card{Class: Door, Color: c})
}

func addDreamCards(d *Deck) {
	for i := 0; i < 10; i++ {
		d.AddCard(&Card{Class: Dream})
	}
}

func (d *Deck) Shuffle() {
	deck := *d
	for i := 0; i < len(deck); i++ {
		j := i + rand.Intn(len(deck)-i)
		deck[i], deck[j] = deck[j], deck[i]
	}
}

func (d *Deck) DrawCard() (*Card, error) {
	if len(*d) == 0 {
		return nil, errors.New("No cards left, you lose.")
	}
	c := (*d)[0]
	*d = (*d)[1:]
	return c, nil
}

func (d *Deck) LastCard() *Card {
	if len(*d) == 0 {
		return nil
	}
	return (*d)[len(*d)-1]
}

func (d *Deck) AddCard(c *Card) {
	*d = append(*d, c)
}

func (d *Deck) RemoveCardAt(i int) *Card {
	deck := *d
	c := deck[i]
	*d = append(deck[:i], deck[i+1:]...)
	return c
}

// FillHand fills the hand up to five cards.
func FillHand(deck, hand *Deck) error {
	var err error
	limbo := make(Deck, 0)
	defer func() {
		if err == nil && len(limbo) > 0 {
			log.Print("Shuffling Limbo into Deck")
			*deck = append(*deck, limbo...)
			(*deck).Shuffle()
		}
	}()

	for {
		if len(*hand) >= 5 {
			break
		}
		c, err := deck.DrawCard()
		if err != nil {
			return err
		}
		if c.Class != Labyrinth {
			log.Printf("%s moved to Limbo", c)
			limbo = append(limbo, c)
			continue
		}
		*hand = append(*hand, c)
		log.Printf("%s added to Hand", c)
	}
	return nil
}

// DrawHand draws a new hand from the deck.
func (d *Deck) DrawHand() (Deck, error) {
	hand := make(Deck, 0)
	if err := FillHand(d, &hand); err != nil {
		return nil, err
	}
	return hand, nil
}

// MakeDeck makes the basic Onirim deck.
func MakeDeck() Deck {
	d := make(Deck, 0, 76)
	addLabyrinthCards(&d, Red, 9)
	addLabyrinthCards(&d, Blue, 8)
	addLabyrinthCards(&d, Green, 7)
	addLabyrinthCards(&d, Brown, 6)
	addDoorCards(&d, Red)
	addDoorCards(&d, Blue)
	addDoorCards(&d, Green)
	addDoorCards(&d, Brown)
	addDreamCards(&d)
	d.Shuffle()
	return d
}
