package main

import (
	"image"
	"image/color"
	"log"
	"math/rand"
	"sort"

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

// Tile type constants
const (
	TileGrass = iota
	TilePath
	TileWater
	TileBridge
	TileMountain
)

// Layer constants
const (
	LayerBase = iota
	LayerOverlay
	LayerObjects
	LayerCount
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
	camera        Camera // Added camera field
}

// Camera tracks the viewport
type Camera struct {
	x, y float32
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
	// Layer the player is currently on (for bridges, etc.)
	currentLayer int
}

// Map represents the game world
type Map struct {
	tiles       [LayerCount][][]int
	width       int
	height      int
	grassTiles  map[string]bool
	bridgeTiles map[string]bool
	// Add collision map
	collisionMap map[string]bool
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
			currentLayer:  LayerBase,
		},
		gameState:     StateOverworld,
		encounterRate: 0.02,
		fontFace:      text.NewGoXFace(basicfont.Face7x13),
		camera: Camera{ // Initialize camera position
			x: 0,
			y: 0,
		},
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

	// Create a map with layers
	game.initMap()

	// Initialize camera to center on player
	game.updateCamera()

	return game
}

// Initialize a map with layers, including more realistic water bodies and bridges
func (g *Game) initMap() {
	width, height := 20, 15
	g.worldMap = Map{
		width:        width,
		height:       height,
		grassTiles:   make(map[string]bool),
		bridgeTiles:  make(map[string]bool),
		collisionMap: make(map[string]bool),
	}

	// Initialize layers
	for layer := 0; layer < LayerCount; layer++ {
		g.worldMap.tiles[layer] = make([][]int, height)
		for y := 0; y < height; y++ {
			g.worldMap.tiles[layer][y] = make([]int, width)
			for x := 0; x < width; x++ {
				g.worldMap.tiles[layer][y][x] = TileGrass // Default to grass

				// Mark as grass tile for encounter checks
				key := formatCoord(x, y)
				g.worldMap.grassTiles[key] = true
			}
		}
	}

	// Generate realistic water bodies using cellular automata
	g.generateWaterBodies(width, height)

	// Generate paths connecting different areas
	g.generatePaths(width, height)

	// Place mountains in clusters away from water
	g.generateMountains(width, height)

	// Add bridges at strategic locations
	g.placeBridges(width, height)
}

// generateWaterBodies creates realistic water features using cellular automata
func (g *Game) generateWaterBodies(width, height int) {
	// Initialize water cells randomly (about 30% of tiles)
	waterMap := make([][]bool, height)
	for y := 0; y < height; y++ {
		waterMap[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			if rand.Float32() < 0.3 {
				waterMap[y][x] = true
			}
		}
	}

	// Run cellular automata iterations to form natural-looking water bodies
	for i := 0; i < 4; i++ {
		newWaterMap := make([][]bool, height)
		for y := 0; y < height; y++ {
			newWaterMap[y] = make([]bool, width)
			for x := 0; x < width; x++ {
				// Count water neighbors (8-way)
				waterNeighbors := 0
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						nx, ny := x+dx, y+dy
						if nx >= 0 && nx < width && ny >= 0 && ny < height && waterMap[ny][nx] {
							waterNeighbors++
						}
					}
				}

				// Apply cellular automata rules:
				// - If a cell has 4+ water neighbors, it becomes water
				// - If a cell has 3 or fewer water neighbors, it becomes land
				newWaterMap[y][x] = waterNeighbors >= 4
			}
		}
		waterMap = newWaterMap
	}

	// Create rivers by drawing lines between water bodies
	riverOrigins := []struct{ x, y int }{}

	// Find potential river origins (water near land)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if waterMap[y][x] {
				hasLandNeighbor := false
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						nx, ny := x+dx, y+dy
						if nx >= 0 && nx < width && ny >= 0 && ny < height && !waterMap[ny][nx] {
							hasLandNeighbor = true
							break
						}
					}
					if hasLandNeighbor {
						break
					}
				}
				if hasLandNeighbor && rand.Float32() < 0.2 {
					riverOrigins = append(riverOrigins, struct{ x, y int }{x, y})
				}
			}
		}
	}

	// Draw rivers from origins
	for _, origin := range riverOrigins {
		if len(riverOrigins) <= 2 || rand.Float32() < 0.5 {
			// Create river path
			x, y := origin.x, origin.y
			length := rand.Intn(8) + 3
			dx, dy := 0, 0

			// Choose a consistent direction for the river
			if rand.Float32() < 0.5 {
				dx = rand.Intn(3) - 1 // -1, 0, or 1
				if dx == 0 {
					dy = rand.Intn(2)*2 - 1 // -1 or 1
				}
			} else {
				dy = rand.Intn(3) - 1 // -1, 0, or 1
				if dy == 0 {
					dx = rand.Intn(2)*2 - 1 // -1 or 1
				}
			}

			// Draw the river
			for i := 0; i < length; i++ {
				nx, ny := x+dx, y+dy
				if nx < 0 || nx >= width || ny < 0 || ny >= height {
					break
				}

				waterMap[ny][nx] = true

				// Slight chance of changing direction
				if rand.Float32() < 0.2 {
					if rand.Float32() < 0.5 {
						dx += rand.Intn(3) - 1
						if dx < -1 {
							dx = -1
						} else if dx > 1 {
							dx = 1
						}
					} else {
						dy += rand.Intn(3) - 1
						if dy < -1 {
							dy = -1
						} else if dy > 1 {
							dy = 1
						}
					}

					// Ensure we have direction
					if dx == 0 && dy == 0 {
						if rand.Float32() < 0.5 {
							dx = rand.Intn(2)*2 - 1
						} else {
							dy = rand.Intn(2)*2 - 1
						}
					}
				}

				x, y = nx, ny
			}
		}
	}

	// Apply water map to the game map
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if waterMap[y][x] {
				g.worldMap.tiles[LayerBase][y][x] = TileWater

				// Add water to collision map
				key := formatCoord(x, y)
				g.worldMap.collisionMap[key] = true
				delete(g.worldMap.grassTiles, key)
			}
		}
	}
}

// generatePaths creates paths connecting different parts of the map
func (g *Game) generatePaths(width, height int) {
	// Create a few random path starting points
	pathPoints := []struct{ x, y int }{}

	// Add a few starting points for paths
	numPathPoints := rand.Intn(3) + 2
	for i := 0; i < numPathPoints; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		pathPoints = append(pathPoints, struct{ x, y int }{x, y})
	}

	// Connect path points with each other
	for i := 0; i < len(pathPoints)-1; i++ {
		start := pathPoints[i]
		end := pathPoints[i+1]

		// Simple pathfinding to connect points
		x, y := start.x, start.y
		for x != end.x || y != end.y {
			if g.worldMap.tiles[LayerBase][y][x] != TileWater {
				g.worldMap.tiles[LayerBase][y][x] = TilePath

				// Remove from grass tiles for encounter checks
				key := formatCoord(x, y)
				delete(g.worldMap.grassTiles, key)
			}

			// Move toward end point
			if x < end.x && rand.Float32() < 0.7 {
				x++
			} else if x > end.x && rand.Float32() < 0.7 {
				x--
			} else if y < end.y {
				y++
			} else if y > end.y {
				y--
			}
		}

		// Set final tile (if not water)
		if g.worldMap.tiles[LayerBase][end.y][end.x] != TileWater {
			g.worldMap.tiles[LayerBase][end.y][end.x] = TilePath

			// Remove from grass tiles for encounter checks
			key := formatCoord(end.x, end.y)
			delete(g.worldMap.grassTiles, key)
		}
	}
}

// generateMountains places mountain clusters in sensible locations
func (g *Game) generateMountains(width, height int) {
	// Add mountains (impassable) in clusters
	numMountainClusters := rand.Intn(3) + 1
	for cluster := 0; cluster < numMountainClusters; cluster++ {
		// Find a spot for mountains (preferably away from water)
		var mountainX, mountainY int
		attempts := 0
		validSpot := false

		for !validSpot && attempts < 20 {
			mountainX = rand.Intn(width-4) + 2
			mountainY = rand.Intn(height-4) + 2

			// Check if the area has minimal water
			waterCount := 0
			for dy := -2; dy <= 2; dy++ {
				for dx := -2; dx <= 2; dx++ {
					nx, ny := mountainX+dx, mountainY+dy
					if nx >= 0 && nx < width && ny >= 0 && ny < height &&
						g.worldMap.tiles[LayerBase][ny][nx] == TileWater {
						waterCount++
					}
				}
			}

			validSpot = waterCount <= 2
			attempts++
		}

		// Create mountain cluster
		clusterSize := rand.Intn(8) + 5
		for i := 0; i < clusterSize; i++ {
			// Mountains form in connected patterns
			offsetX := rand.Intn(5) - 2
			offsetY := rand.Intn(5) - 2

			nx, ny := mountainX+offsetX, mountainY+offsetY
			if nx >= 0 && nx < width && ny >= 0 && ny < height &&
				g.worldMap.tiles[LayerBase][ny][nx] != TileWater {
				g.worldMap.tiles[LayerBase][ny][nx] = TileMountain

				// Add mountain to collision map
				key := formatCoord(nx, ny)
				g.worldMap.collisionMap[key] = true
				delete(g.worldMap.grassTiles, key)
			}
		}
	}
}

// placeBridges adds bridges at strategic locations over water
func (g *Game) placeBridges(width, height int) {
	// Find potential bridge locations by looking for water bodies that separate land
	bridgeCandidates := []struct {
		x, y      int
		direction int // 0 for horizontal, 1 for vertical
		length    int // Length of water to cross
	}{}

	// Find horizontal bridge candidates
	for y := 1; y < height-1; y++ {
		for x := 1; x < width-2; x++ {
			// Look for patterns like: land - water - water - land
			if g.worldMap.tiles[LayerBase][y][x-1] != TileWater &&
				g.worldMap.tiles[LayerBase][y][x] == TileWater {

				// Find the end of the water stretch
				endX := x
				for endX < width-1 && g.worldMap.tiles[LayerBase][y][endX] == TileWater {
					endX++
				}

				// If we found land on the other side and water stretch isn't too long
				// but also not too short (at least 2 tiles of water)
				waterLength := endX - x
				if endX < width &&
					g.worldMap.tiles[LayerBase][y][endX] != TileWater &&
					waterLength >= 2 && waterLength <= 5 {

					// Check that this isn't just following the coastline
					// by ensuring both sides are actual separate land masses

					// Check left side isn't just a thin peninsula
					leftIsSolid := false
					if x-1 >= 0 && y-1 >= 0 && y+1 < height {
						landCount := 0
						if g.worldMap.tiles[LayerBase][y-1][x-1] != TileWater {
							landCount++
						}
						if g.worldMap.tiles[LayerBase][y+1][x-1] != TileWater {
							landCount++
						}
						leftIsSolid = landCount >= 1
					}

					// Check right side isn't just a thin peninsula
					rightIsSolid := false
					if endX < width && y-1 >= 0 && y+1 < height {
						landCount := 0
						if g.worldMap.tiles[LayerBase][y-1][endX] != TileWater {
							landCount++
						}
						if g.worldMap.tiles[LayerBase][y+1][endX] != TileWater {
							landCount++
						}
						rightIsSolid = landCount >= 1
					}

					// Only add if bridge connects solid land on both sides
					if leftIsSolid && rightIsSolid {
						bridgeCandidates = append(bridgeCandidates, struct {
							x, y, direction, length int
						}{x, y, 0, waterLength})
					}
				}
			}
		}
	}

	// Find vertical bridge candidates
	for x := 1; x < width-1; x++ {
		for y := 1; y < height-2; y++ {
			// Look for patterns like: land - water - water - land
			if g.worldMap.tiles[LayerBase][y-1][x] != TileWater &&
				g.worldMap.tiles[LayerBase][y][x] == TileWater {

				// Find the end of the water stretch
				endY := y
				for endY < height-1 && g.worldMap.tiles[LayerBase][endY][x] == TileWater {
					endY++
				}

				// If we found land on the other side and water stretch isn't too long
				// but also not too short (at least 2 tiles of water)
				waterLength := endY - y
				if endY < height &&
					g.worldMap.tiles[LayerBase][endY][x] != TileWater &&
					waterLength >= 2 && waterLength <= 5 {

					// Check that this isn't just following the coastline
					// by ensuring both sides are actual separate land masses

					// Check top side isn't just a thin peninsula
					topIsSolid := false
					if y-1 >= 0 && x-1 >= 0 && x+1 < width {
						landCount := 0
						if g.worldMap.tiles[LayerBase][y-1][x-1] != TileWater {
							landCount++
						}
						if g.worldMap.tiles[LayerBase][y-1][x+1] != TileWater {
							landCount++
						}
						topIsSolid = landCount >= 1
					}

					// Check bottom side isn't just a thin peninsula
					bottomIsSolid := false
					if endY < height && x-1 >= 0 && x+1 < width {
						landCount := 0
						if g.worldMap.tiles[LayerBase][endY][x-1] != TileWater {
							landCount++
						}
						if g.worldMap.tiles[LayerBase][endY][x+1] != TileWater {
							landCount++
						}
						bottomIsSolid = landCount >= 1
					}

					// Only add if bridge connects solid land on both sides
					if topIsSolid && bottomIsSolid {
						bridgeCandidates = append(bridgeCandidates, struct {
							x, y, direction, length int
						}{x, y, 1, waterLength})
					}
				}
			}
		}
	}

	// Score and sort bridge candidates
	type scoredBridge struct {
		x, y      int
		direction int
		length    int
		score     int
	}

	scoredBridges := make([]scoredBridge, 0, len(bridgeCandidates))
	for _, bridge := range bridgeCandidates {
		// Calculate a score based on:
		// 1. Prefer bridges that cross larger stretches of water (but not too large)
		// 2. Prefer bridges that are in more central map positions
		lengthScore := bridge.length * 10

		// Calculate distance from center of map
		centerX, centerY := width/2, height/2
		distX := abs(bridge.x - centerX)
		distY := abs(bridge.y - centerY)
		centralityScore := 20 - (distX + distY)
		if centralityScore < 0 {
			centralityScore = 0
		}

		totalScore := lengthScore + centralityScore

		scoredBridges = append(scoredBridges, scoredBridge{
			x:         bridge.x,
			y:         bridge.y,
			direction: bridge.direction,
			length:    bridge.length,
			score:     totalScore,
		})
	}

	// Sort bridges by score (highest first)
	sort.Slice(scoredBridges, func(i, j int) bool {
		return scoredBridges[i].score > scoredBridges[j].score
	})

	// Place up to 3 bridges (if we have that many candidates)
	numBridges := min(len(scoredBridges), 3)

	// Keep track of bridge locations to avoid building bridges too close together
	bridgeMap := make(map[string]bool)

	// Place highest scoring bridges
	bridgesPlaced := 0
	for i := 0; i < len(scoredBridges) && bridgesPlaced < numBridges; i++ {
		bridge := scoredBridges[i]

		// Check if this bridge is too close to an existing bridge
		tooClose := false

		if bridge.direction == 0 { // Horizontal bridge
			// Find end of water
			endX := bridge.x
			for endX < width && g.worldMap.tiles[LayerBase][bridge.y][endX] == TileWater {
				endX++
			}

			// Check proximity to other bridges
			buffer := 2 // Minimum distance between bridges
			for y := bridge.y - buffer; y <= bridge.y+buffer; y++ {
				for x := bridge.x - buffer; x <= endX+buffer; x++ {
					key := formatCoord(x, y)
					if bridgeMap[key] {
						tooClose = true
						break
					}
				}
				if tooClose {
					break
				}
			}

			if !tooClose {
				// Place bridge tiles over water
				for x := bridge.x; x < endX; x++ {
					g.worldMap.tiles[LayerOverlay][bridge.y][x] = TileBridge
					key := formatCoord(x, bridge.y)
					g.worldMap.bridgeTiles[key] = true
					delete(g.worldMap.collisionMap, key)
					bridgeMap[key] = true
				}
				bridgesPlaced++
			}
		} else { // Vertical bridge
			// Find end of water
			endY := bridge.y
			for endY < height && g.worldMap.tiles[LayerBase][endY][bridge.x] == TileWater {
				endY++
			}

			// Check proximity to other bridges
			buffer := 2 // Minimum distance between bridges
			for y := bridge.y - buffer; y <= endY+buffer; y++ {
				for x := bridge.x - buffer; x <= bridge.x+buffer; x++ {
					key := formatCoord(x, y)
					if bridgeMap[key] {
						tooClose = true
						break
					}
				}
				if tooClose {
					break
				}
			}

			if !tooClose {
				// Place bridge tiles over water
				for y := bridge.y; y < endY; y++ {
					g.worldMap.tiles[LayerOverlay][y][bridge.x] = TileBridge
					key := formatCoord(bridge.x, y)
					g.worldMap.bridgeTiles[key] = true
					delete(g.worldMap.collisionMap, key)
					bridgeMap[key] = true
				}
				bridgesPlaced++
			}
		}
	}
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// min returns the smaller of a and b
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function to format coordinates for the various tile maps
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

// updateCamera centers the camera on the player with smooth movement
func (g *Game) updateCamera() {
	// Calculate the target camera position (centered on player)
	targetX := g.player.visualX - screenWidth/2 + tileSize/2
	targetY := g.player.visualY - screenHeight/2 + tileSize/2

	// Smoothly move camera towards target (can adjust the 0.1 for different smoothing)
	g.camera.x += (targetX - g.camera.x) * 0.1
	g.camera.y += (targetY - g.camera.y) * 0.1

	// Clamp camera to map bounds
	if g.camera.x < 0 {
		g.camera.x = 0
	}
	if g.camera.y < 0 {
		g.camera.y = 0
	}
	maxX := float32(g.worldMap.width*tileSize) - screenWidth
	maxY := float32(g.worldMap.height*tileSize) - screenHeight
	if g.camera.x > maxX {
		g.camera.x = maxX
	}
	if g.camera.y > maxY {
		g.camera.y = maxY
	}
}

// updateOverworld handles overworld state updates
func (g *Game) updateOverworld() {
	// Handle movement based on the current state
	switch g.player.movementState {
	case MovementIdle:
		// Check for key presses for continuous movement
		g.handlePlayerMovement()

	case MovementMoving:
		// Update visual position to smoothly move toward the target tile
		targetX := float32(g.player.tileX * tileSize)
		targetY := float32(g.player.tileY * tileSize)

		// Calculate how fast to move
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

			// Check for bridge tiles and adjust player layer
			key := formatCoord(g.player.tileX, g.player.tileY)
			if g.worldMap.bridgeTiles[key] {
				g.player.currentLayer = LayerOverlay
			} else {
				g.player.currentLayer = LayerBase
			}

			// Check for wild creature encounters in grass when arriving at a new tile
			if g.worldMap.grassTiles[key] && g.player.currentLayer == LayerBase && rand.Float32() < g.encounterRate {
				g.startBattle()
			}

			// Continue movement if key is still held (for continuous movement)
			g.handlePlayerMovement()
		}
	}

	// Update camera position to follow player
	g.updateCamera()
}

// handlePlayerMovement processes player movement input
func (g *Game) handlePlayerMovement() {
	// Variable to track if we've started movement
	moved := false

	// Handle arrow keys for movement
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.player.direction = DirectionUp
		// Check if we can move to the target tile
		newY := g.player.tileY - 1
		if newY >= 0 && !g.isCollision(g.player.tileX, newY) {
			g.player.tileY = newY
			moved = true
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.player.direction = DirectionDown
		// Check if we can move to the target tile
		newY := g.player.tileY + 1
		if newY < g.worldMap.height && !g.isCollision(g.player.tileX, newY) {
			g.player.tileY = newY
			moved = true
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.direction = DirectionLeft
		// Check if we can move to the target tile
		newX := g.player.tileX - 1
		if newX >= 0 && !g.isCollision(newX, g.player.tileY) {
			g.player.tileX = newX
			moved = true
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.direction = DirectionRight
		// Check if we can move to the target tile
		newX := g.player.tileX + 1
		if newX < g.worldMap.width && !g.isCollision(newX, g.player.tileY) {
			g.player.tileX = newX
			moved = true
		}
	}

	// If we moved, update the movement state
	if moved {
		g.player.movementState = MovementMoving
	}
}

// isCollision checks if a tile is impassable
func (g *Game) isCollision(x, y int) bool {
	key := formatCoord(x, y)
	return g.worldMap.collisionMap[key]
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
	// Draw the base layer first
	g.drawMapLayer(screen, LayerBase)

	// Draw the overlay layer (bridges, etc.)
	g.drawMapLayer(screen, LayerOverlay)

	// Draw the player at visual position (for smooth movement)
	playerColor := color.RGBA{255, 0, 0, 255}
	vector.DrawFilledRect(
		screen,
		g.player.visualX-g.camera.x,
		g.player.visualY-g.camera.y,
		tileSize,
		tileSize,
		playerColor,
		true,
	)

	// Draw player direction indicator
	indicatorSize := tileSize / 4

	switch g.player.direction {
	case DirectionUp: // Up
		vector.DrawFilledRect(
			screen,
			g.player.visualX-g.camera.x+float32(tileSize/2-indicatorSize/2),
			g.player.visualY-g.camera.y,
			float32(indicatorSize),
			float32(indicatorSize),
			color.White,
			true,
		)
	case DirectionDown: // Down
		vector.DrawFilledRect(
			screen,
			g.player.visualX-g.camera.x+float32(tileSize/2-indicatorSize/2),
			g.player.visualY-g.camera.y+float32(tileSize-indicatorSize),
			float32(indicatorSize),
			float32(indicatorSize),
			color.White,
			true,
		)
	case DirectionLeft: // Left
		vector.DrawFilledRect(
			screen,
			g.player.visualX-g.camera.x,
			g.player.visualY-g.camera.y+float32(tileSize/2-indicatorSize/2),
			float32(indicatorSize),
			float32(indicatorSize),
			color.White,
			true,
		)
	case DirectionRight: // Right
		vector.DrawFilledRect(
			screen,
			g.player.visualX-g.camera.x+float32(tileSize-indicatorSize),
			g.player.visualY-g.camera.y+float32(tileSize/2-indicatorSize/2),
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
	// text.Draw(screen, fmt.Sprintf("Tile: %d,%d Layer: %d", g.player.tileX, g.player.tileY, g.player.currentLayer), g.fontFace, op)
}

// drawMapLayer draws a specific layer of the map
func (g *Game) drawMapLayer(screen *ebiten.Image, layer int) {
	// Calculate visible tile range based on camera position
	startX := int(g.camera.x) / tileSize
	startY := int(g.camera.y) / tileSize
	endX := startX + screenWidth/tileSize + 2 // +2 to handle partially visible tiles
	endY := startY + screenHeight/tileSize + 2

	// Clamp to map bounds
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}
	if endX > g.worldMap.width {
		endX = g.worldMap.width
	}
	if endY > g.worldMap.height {
		endY = g.worldMap.height
	}

	// Only draw visible tiles
	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			tile := g.worldMap.tiles[layer][y][x]
			if tile == 0 && layer > LayerBase {
				continue // Skip empty tiles in overlay layers
			}

			var tileColor color.RGBA

			switch tile {
			case TileGrass:
				tileColor = color.RGBA{34, 139, 34, 255} // Green
			case TilePath:
				tileColor = color.RGBA{210, 180, 140, 255} // Brown
			case TileWater:
				tileColor = color.RGBA{30, 144, 255, 255} // Blue
			case TileBridge:
				tileColor = color.RGBA{139, 69, 19, 255} // Dark brown
			case TileMountain:
				tileColor = color.RGBA{105, 105, 105, 255} // Dark grey
			default:
				continue // Skip drawing if empty
			}

			vector.DrawFilledRect(
				screen,
				float32(x*tileSize)-g.camera.x,
				float32(y*tileSize)-g.camera.y,
				tileSize,
				tileSize,
				tileColor,
				true,
			)
		}
	}
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
	ebiten.SetWindowTitle("Creaturegame")

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
