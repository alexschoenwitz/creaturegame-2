package main

import (
	"image"
	"image/color"
)

// Creature represents a creature in the game
type Creature struct {
	name     string
	hp       int
	maxHP    int
	attack   int
	defense  int
	speed    int
	type1    string
	moves    []Move
	level    int
	inBattle bool
	position image.Point
	color    color.RGBA
}

// Move represents a move/attack
type Move struct {
	name     string
	power    int
	accuracy int
	type1    string
}
