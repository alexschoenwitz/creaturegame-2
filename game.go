package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/basicfont"
)

// Game state constants
const (
	StateMainMenu = iota
	StateOverworld
	StateBattle
	StateMenu
)

// Game is the main game struct
type Game struct {
	player          Player
	gameState       int
	worldMap        Map
	battle          Battle
	encounterRate   float32
	creatures       []Creature
	fontFace        text.Face
	camera          Camera
	menuOptions     []string
	selectedOption  int
	gameInitialized bool
}

// NewGame creates a new game instance
func NewGame() *Game {
	game := &Game{
		player: Player{
			tileX:         5,
			tileY:         5,
			visualX:       float32(5 * tileSize),
			visualY:       float32(5 * tileSize),
			movementState: MovementIdle,
			direction:     DirectionDown,
			currentLayer:  LayerBase,
		},
		gameState:     StateMainMenu, // Start with main menu
		encounterRate: 0.02,
		fontFace:      text.NewGoXFace(basicfont.Face7x13),
		camera: Camera{
			x: 0,
			y: 0,
		},
		menuOptions:     []string{"New Game", "Options", "Exit"},
		selectedOption:  0,
		gameInitialized: false,
	}

	game.initGame()

	return game
}

// initGame initializes the game world and creatures
func (g *Game) initGame() {
	if g.gameInitialized {
		return
	}

	// Create creatures
	// Create some creatures
	g.creatures = []Creature{
		{
			name:     "Sparkitty",
			hp:       50,
			maxHP:    50,
			attack:   12,
			defense:  10,
			speed:    15,
			type1:    "Electric",
			level:    5,
			inBattle: false,
			color:    color.RGBA{255, 255, 0, 255},
			moves: []Move{
				{name: "Tackle", power: 40, accuracy: 100, type1: "Normal"},
				{name: "Spark", power: 50, accuracy: 90, type1: "Electric"},
			},
		},
		{
			name:     "Flamepup",
			hp:       45,
			maxHP:    45,
			attack:   15,
			defense:  8,
			speed:    12,
			type1:    "Fire",
			level:    5,
			inBattle: false,
			color:    color.RGBA{255, 100, 0, 255},
			moves: []Move{
				{name: "Tackle", power: 40, accuracy: 100, type1: "Normal"},
				{name: "Ember", power: 50, accuracy: 90, type1: "Fire"},
			},
		},
		{
			name:     "Bubblefrog",
			hp:       55,
			maxHP:    55,
			attack:   10,
			defense:  12,
			speed:    10,
			type1:    "Water",
			level:    5,
			inBattle: false,
			color:    color.RGBA{0, 100, 255, 255},
			moves: []Move{
				{name: "Tackle", power: 40, accuracy: 100, type1: "Normal"},
				{name: "Bubble", power: 50, accuracy: 90, type1: "Water"},
			},
		},
	}

	// Initialize the player's starter creature
	g.battle.playerCreature = g.creatures[0]

	// Create the map with layers
	g.initMap()

	// Initialize camera to center on player
	g.updateCamera()

	g.gameInitialized = true
}

// Update updates the game state
func (g *Game) Update() error {
	switch g.gameState {
	case StateMainMenu:
		g.updateMainMenu()
	case StateOverworld:
		g.updateOverworld()
	case StateBattle:
		g.updateBattle()
	}
	return nil
}

// Draw draws the game
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear the screen
	screen.Fill(color.RGBA{135, 206, 235, 255})

	switch g.gameState {
	case StateMainMenu:
		g.drawMainMenu(screen)
	case StateOverworld:
		g.drawOverworld(screen)
	case StateBattle:
		g.drawBattle(screen)
	}
}

// Layout implements ebiten.Game's Layout
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
