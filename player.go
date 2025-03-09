package main

import "github.com/hajimehoshi/ebiten/v2"

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
