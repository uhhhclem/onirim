package main

import (
	"fmt"
	"onirim"
)

func main() {

	d := onirim.MakeDeck()
	h, err := d.DrawHand()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Hand: %s", h)
}
