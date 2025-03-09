package main

import (
	"image"
	"image/color"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

const (
	screenWidth  = 320
	screenHeight = 240
	tileSize     = 32
)

// Game state constants
const (
	StateOverworld = iota
	StateBattle
	StateMenu
)

// Movement states for tile-based movement
const (
	MovementIdle = iota
	MovementMoving
)

// Direction constants
const (
	DirectionUp = iota
	DirectionDown
	DirectionLeft
	DirectionRight
)

// Game is the main game struct
type Game struct {
	player        Player
	gameState     int
	worldMap      Map
	battle        Battle
	encounterRate float32
	creatures     []Creature
	fontFace      text.Face
}

// Player represents the player's character
type Player struct {
	// Current position in tiles
	tileX, tileY int
	// Visual position in pixels for smooth movement
	visualX, visualY float32
	// Movement state tracking
	movementState int
	direction     int
	frameCount    int
}

// Map represents the game world
type Map struct {
	tiles      [][]int
	width      int
	height     int
	grassTiles map[string]bool
}

// Battle represents a battle state
type Battle struct {
	playerCreature  Creature
	enemyCreature   Creature
	currentTurn     int
	selectedAction  int
	battleText      string
	battleTextTimer int
}

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
		},
		gameState:     StateOverworld,
		encounterRate: 0.02,
		fontFace:      text.NewGoXFace(basicfont.Face7x13),
	}

	// Create some creatures
	game.creatures = []Creature{
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
	game.battle.playerCreature = game.creatures[0]

	// Create a simple map
	game.initMap()

	return game
}

// Initialize a simple tile map
func (g *Game) initMap() {
	width, height := 20, 15
	g.worldMap = Map{
		tiles:      make([][]int, height),
		width:      width,
		height:     height,
		grassTiles: make(map[string]bool),
	}

	// Generate a simple map with grass and path
	for y := range height {
		g.worldMap.tiles[y] = make([]int, width)
		for x := range width {
			// 0 = path, 1 = grass
			if rand.Float32() < 0.7 {
				g.worldMap.tiles[y][x] = 1
				// Mark as grass tile for encounter checks
				key := formatCoord(x, y)
				g.worldMap.grassTiles[key] = true
			} else {
				g.worldMap.tiles[y][x] = 0
			}
		}
	}
}

// Helper function to format coordinates for the grass tiles map
func formatCoord(x, y int) string {
	return string(rune(x)) + "," + string(rune(y))
}

// Update updates the game state
func (g *Game) Update() error {
	switch g.gameState {
	case StateOverworld:
		g.updateOverworld()
	case StateBattle:
		g.updateBattle()
	}
	return nil
}

// updateOverworld handles overworld state updates
func (g *Game) updateOverworld() {
	// Handle movement based on the current state
	switch g.player.movementState {
	case MovementIdle:
		// Player is not moving, check for input
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
			g.player.direction = DirectionUp
			// Check if we can move to the target tile
			if g.player.tileY > 0 {
				g.player.tileY--
				g.player.movementState = MovementMoving
			}
		} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
			g.player.direction = DirectionDown
			// Check if we can move to the target tile
			if g.player.tileY < g.worldMap.height-1 {
				g.player.tileY++
				g.player.movementState = MovementMoving
			}
		} else if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
			g.player.direction = DirectionLeft
			// Check if we can move to the target tile
			if g.player.tileX > 0 {
				g.player.tileX--
				g.player.movementState = MovementMoving
			}
		} else if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
			g.player.direction = DirectionRight
			// Check if we can move to the target tile
			if g.player.tileX < g.worldMap.width-1 {
				g.player.tileX++
				g.player.movementState = MovementMoving
			}
		}

	case MovementMoving:
		// Update visual position to smoothly move toward the target tile
		targetX := float32(g.player.tileX * tileSize)
		targetY := float32(g.player.tileY * tileSize)

		// Calculate how fast to move (adjust the divisor to change speed)
		const movementSpeed = 4.0

		// Update visual position
		if g.player.visualX < targetX {
			g.player.visualX += movementSpeed
			if g.player.visualX > targetX {
				g.player.visualX = targetX
			}
		} else if g.player.visualX > targetX {
			g.player.visualX -= movementSpeed
			if g.player.visualX < targetX {
				g.player.visualX = targetX
			}
		}

		if g.player.visualY < targetY {
			g.player.visualY += movementSpeed
			if g.player.visualY > targetY {
				g.player.visualY = targetY
			}
		} else if g.player.visualY > targetY {
			g.player.visualY -= movementSpeed
			if g.player.visualY < targetY {
				g.player.visualY = targetY
			}
		}

		// Animation frame count
		g.player.frameCount++

		// Check if movement is complete
		if g.player.visualX == targetX && g.player.visualY == targetY {
			g.player.movementState = MovementIdle

			// Check for wild creature encounters in grass when arriving at a new tile
			key := formatCoord(g.player.tileX, g.player.tileY)
			if g.worldMap.grassTiles[key] && rand.Float32() < g.encounterRate {
				g.startBattle()
			}
		}
	}
}

// Start a battle with a random wild creature
func (g *Game) startBattle() {
	g.gameState = StateBattle

	// Select a random creature as the enemy
	enemyIndex := rand.Intn(len(g.creatures))
	g.battle.enemyCreature = g.creatures[enemyIndex]

	// Reset the creature's HP for the battle
	g.battle.enemyCreature.hp = g.battle.enemyCreature.maxHP

	// Set up the battle state
	g.battle.currentTurn = 0
	g.battle.selectedAction = 0
	g.battle.battleText = "A wild " + g.battle.enemyCreature.name + " appeared!"
	g.battle.battleTextTimer = 60 // Show text for 60 frames
}

// updateBattle handles battle state updates
func (g *Game) updateBattle() {
	// Update battle text timer
	if g.battle.battleTextTimer > 0 {
		g.battle.battleTextTimer--
		return
	}

	// Handle player input during battle
	if g.battle.currentTurn == 0 {
		// Player's turn
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
			g.battle.selectedAction = (g.battle.selectedAction - 1 + len(g.battle.playerCreature.moves)) % len(g.battle.playerCreature.moves)
		} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
			g.battle.selectedAction = (g.battle.selectedAction + 1) % len(g.battle.playerCreature.moves)
		}

		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			// Execute selected move
			selectedMove := g.battle.playerCreature.moves[g.battle.selectedAction]
			damage := calculateDamage(g.battle.playerCreature, g.battle.enemyCreature, selectedMove)

			g.battle.enemyCreature.hp -= damage
			if g.battle.enemyCreature.hp < 0 {
				g.battle.enemyCreature.hp = 0
			}

			g.battle.battleText = g.battle.playerCreature.name + " used " + selectedMove.name + "!"
			g.battle.battleTextTimer = 60
			g.battle.currentTurn = 1 // Switch to enemy turn
		}
	} else {
		// Enemy's turn
		if g.battle.battleTextTimer <= 0 {
			if g.battle.enemyCreature.hp <= 0 {
				g.battle.battleText = g.battle.enemyCreature.name + " fainted!"
				g.battle.battleTextTimer = 60
				g.gameState = StateOverworld
			} else {
				// Enemy attacks with a random move
				enemyMoveIndex := rand.Intn(len(g.battle.enemyCreature.moves))
				enemyMove := g.battle.enemyCreature.moves[enemyMoveIndex]

				damage := calculateDamage(g.battle.enemyCreature, g.battle.playerCreature, enemyMove)

				g.battle.playerCreature.hp -= damage
				if g.battle.playerCreature.hp < 0 {
					g.battle.playerCreature.hp = 0
				}

				g.battle.battleText = g.battle.enemyCreature.name + " used " + enemyMove.name + "!"
				g.battle.battleTextTimer = 60

				if g.battle.playerCreature.hp <= 0 {
					g.battle.battleText = g.battle.playerCreature.name + " fainted!"
					g.battle.battleTextTimer = 60
					g.gameState = StateOverworld

					// Heal player's creature for the next battle
					g.battle.playerCreature.hp = g.battle.playerCreature.maxHP
				} else {
					g.battle.currentTurn = 0 // Switch back to player's turn
				}
			}
		}
	}

	// Check for escape
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.battle.battleText = "Got away safely!"
		g.battle.battleTextTimer = 60
		g.gameState = StateOverworld
	}
}

// calculateDamage calculates damage from an attack
func calculateDamage(attacker, defender Creature, move Move) int {
	// Basic damage formula similar to Pokémon
	baseDamage := (2*attacker.level)/5 + 2
	baseDamage = baseDamage * move.power * attacker.attack / defender.defense
	baseDamage = baseDamage/50 + 2

	// Random factor between 0.85 and 1.0
	randomFactor := 0.85 + rand.Float32()*0.15

	return int(float32(baseDamage) * randomFactor)
}

// Draw draws the game
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear the screen
	screen.Fill(color.RGBA{135, 206, 235, 255})

	switch g.gameState {
	case StateOverworld:
		g.drawOverworld(screen)
	case StateBattle:
		g.drawBattle(screen)
	}
}

// drawOverworld draws the overworld map and player
func (g *Game) drawOverworld(screen *ebiten.Image) {
	// Draw the map
	for y := range g.worldMap.height {
		for x := range g.worldMap.width {
			tile := g.worldMap.tiles[y][x]
			if tile == 0 {
				// Path tile (brown)
				vector.DrawFilledRect(screen, float32(x*tileSize), float32(y*tileSize), tileSize, tileSize, color.RGBA{210, 180, 140, 255}, true)
			} else {
				// Grass tile (green)
				vector.DrawFilledRect(screen, float32(x*tileSize), float32(y*tileSize), tileSize, tileSize, color.RGBA{34, 139, 34, 255}, true)
			}
		}
	}

	// Draw the player at visual position (for smooth movement)
	playerColor := color.RGBA{255, 0, 0, 255}
	vector.DrawFilledRect(screen, g.player.visualX, g.player.visualY, tileSize, tileSize, playerColor, true)

	// Draw player direction indicator
	indicatorSize := tileSize / 4

	switch g.player.direction {
	case DirectionUp: // Up
		vector.DrawFilledRect(
			screen,
			g.player.visualX+float32(tileSize/2-indicatorSize/2),
			g.player.visualY,
			float32(indicatorSize),
			float32(indicatorSize),
			color.White,
			true,
		)
	case DirectionDown: // Down
		vector.DrawFilledRect(
			screen,
			g.player.visualX+float32(tileSize/2-indicatorSize/2),
			g.player.visualY+float32(tileSize-indicatorSize),
			float32(indicatorSize),
			float32(indicatorSize),
			color.White,
			true,
		)
	case DirectionLeft: // Left
		vector.DrawFilledRect(
			screen,
			g.player.visualX,
			g.player.visualY+float32(tileSize/2-indicatorSize/2),
			float32(indicatorSize),
			float32(indicatorSize),
			color.White,
			true,
		)
	case DirectionRight: // Right
		vector.DrawFilledRect(
			screen,
			g.player.visualX+float32(tileSize-indicatorSize),
			g.player.visualY+float32(tileSize/2-indicatorSize/2),
			float32(indicatorSize),
			float32(indicatorSize),
			color.White,
			true,
		)
	}

	// Debug info (optional)
	// op := &text.DrawOptions{}
	// op.GeoM.Translate(10, 10)
	// op.ColorScale.ScaleWithColor(color.White)
	// text.Draw(screen, fmt.Sprintf("Tile: %d,%d", g.player.tileX, g.player.tileY), g.fontFace, op)
}

// drawBattle draws the battle screen
func (g *Game) drawBattle(screen *ebiten.Image) {
	// Draw battle background
	screen.Fill(color.RGBA{200, 200, 200, 255})

	// Draw enemy creature
	enemySize := 40
	enemyX := screenWidth/2 - enemySize/2
	enemyY := 50
	vector.DrawFilledRect(screen, float32(enemyX), float32(enemyY), float32(enemySize), float32(enemySize), g.battle.enemyCreature.color, true)

	// Draw player creature
	playerSize := 40
	playerX := 50
	playerY := screenHeight - 100
	vector.DrawFilledRect(screen, float32(playerX), float32(playerY), float32(playerSize), float32(playerSize), g.battle.playerCreature.color, true)

	// Draw battle UI
	uiRect := image.Rect(0, screenHeight-70, screenWidth, screenHeight)
	vector.DrawFilledRect(screen, float32(uiRect.Min.X), float32(uiRect.Min.Y), float32(uiRect.Dx()), float32(uiRect.Dy()), color.RGBA{50, 50, 50, 240}, true)

	// Draw battle text
	if g.battle.battleTextTimer > 0 {
		op := &text.DrawOptions{}
		op.GeoM.Translate(10, float64(screenHeight-50))
		op.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, g.battle.battleText, g.fontFace, op)
	} else if g.battle.currentTurn == 0 {
		op := &text.DrawOptions{}
		op.GeoM.Translate(10, float64(screenHeight-50))
		op.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, "What will "+g.battle.playerCreature.name+" do?", g.fontFace, op)

		// Draw move options
		for i, move := range g.battle.playerCreature.moves {
			op := &text.DrawOptions{}
			op.GeoM.Translate(30, float64(screenHeight-30+i*15))
			op.ColorScale.ScaleWithColor(color.White)
			text.Draw(screen, move.name, g.fontFace, op)

			op2 := &text.DrawOptions{}
			op2.GeoM.Translate(15, float64(screenHeight-30+i*15))
			op2.ColorScale.ScaleWithColor(color.White)
			// Highlight selected move
			if i == g.battle.selectedAction {
				text.Draw(screen, ">", g.fontFace, op2)
			}
		}
	}

	// Draw HP bars
	// Enemy HP
	vector.DrawFilledRect(screen, float32(enemyX), float32(enemyY-15), float32(enemySize), 5, color.RGBA{100, 100, 100, 255}, true)
	hpRatio := float32(g.battle.enemyCreature.hp) / float32(g.battle.enemyCreature.maxHP)
	hpColor := color.RGBA{0, 255, 0, 255}
	if hpRatio < 0.5 {
		hpColor = color.RGBA{255, 255, 0, 255}
	}
	if hpRatio < 0.2 {
		hpColor = color.RGBA{255, 0, 0, 255}
	}
	vector.DrawFilledRect(screen, float32(enemyX), float32(enemyY-15), float32(enemySize)*hpRatio, 5, hpColor, true)
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(enemyX), float64(enemyY-25))
	op.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, g.battle.enemyCreature.name+" Lv."+string(rune(g.battle.enemyCreature.level+'0')), g.fontFace, op)

	// Player HP
	vector.DrawFilledRect(screen, float32(playerX), float32(playerY-15), float32(playerSize), 5, color.RGBA{100, 100, 100, 255}, true)
	hpRatio = float32(g.battle.playerCreature.hp) / float32(g.battle.playerCreature.maxHP)
	hpColor = color.RGBA{0, 255, 0, 255}
	if hpRatio < 0.5 {
		hpColor = color.RGBA{255, 255, 0, 255}
	}
	if hpRatio < 0.2 {
		hpColor = color.RGBA{255, 0, 0, 255}
	}
	vector.DrawFilledRect(screen, float32(playerX), float32(playerY-15), float32(playerSize)*hpRatio, 5, hpColor, true)
	op2 := &text.DrawOptions{}
	op2.GeoM.Translate(float64(playerX), float64(playerY-25))
	op2.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, g.battle.playerCreature.name+" Lv."+string(rune(g.battle.playerCreature.level+'0')), g.fontFace, op2)
}

// Layout implements ebiten.Game's Layout
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Pokémon Emerald Clone")

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
