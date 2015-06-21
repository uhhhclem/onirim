// Package onirim contains logic for playing Shadi Torbey's game Onirim.
package onirim

import (
	"errors"
	"fmt"
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
	Brown
	Green
	Blue
)

// Colors contains the names of the colors.
var colors = map[ColorEnum]string{
	NoColor: "None",
	Red:     "Red",
	Green:   "Green",
	Brown:   "Brown",
	Blue:    "Blue",
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
	if c.Class == Dream {
		return "[Nightmare]"
	}
	name := "Door"
	if c.Class != Door {
		name = c.Symbol.String()
	}
	return fmt.Sprintf("\033[%dm[%s]\033[37m", c.Color+30, name)
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
