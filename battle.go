package main

import (
	"image"
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Battle represents a battle state
type Battle struct {
	playerCreature  Creature
	enemyCreature   Creature
	currentTurn     int
	selectedAction  int
	battleText      string
	battleTextTimer int
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
