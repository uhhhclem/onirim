package main

import (
	"fmt"
	"onirim"
)

func main() {

	g, err := onirim.NewGame()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Hand: %s", g.Hand)
}
