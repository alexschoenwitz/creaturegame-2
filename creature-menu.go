package main

import (
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// updateCreatureMenu handles updates for the creature management menu
func (g *Game) updateCreatureMenu() {
	if g.menuSection == 0 {
		// In the creature list section
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
			g.selectedCreature = (g.selectedCreature - 1)
			if g.selectedCreature < 0 {
				g.selectedCreature = len(g.creatures) - 1
			}
		} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
			g.selectedCreature = (g.selectedCreature + 1) % len(g.creatures)
		}

		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.menuSection = 1 // Go to detail view for the selected creature
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.gameState = StateOverworld // Return to game
		}
	} else if g.menuSection == 1 {
		// In the creature detail section
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
			g.selectedOption = (g.selectedOption - 1)
			if g.selectedOption < 0 {
				g.selectedOption = len(g.creatureMenuOptions) - 1
			}
		} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
			g.selectedOption = (g.selectedOption + 1) % len(g.creatureMenuOptions)
		}

		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			switch g.selectedOption {
			case 0: // View Stats - already showing
				// Could add more detailed stats in the future
			case 1: // Switch Order
				// If player has more than one creature, allow switching
				if len(g.creatures) > 1 {
					// Update player's main creature
					g.battle.playerCreature = g.creatures[g.selectedCreature]
				}
			case 2: // Back
				g.menuSection = 0 // Return to creature list
				g.selectedOption = 0
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.menuSection = 0 // Return to creature list
			g.selectedOption = 0
		}
	}
} // drawCreatureMenu draws the creature management menu
func (g *Game) drawCreatureMenu(screen *ebiten.Image) {
	// Draw the menu background
	vector.DrawFilledRect(
		screen,
		10,
		10,
		float32(screenWidth-20),
		float32(screenHeight-20),
		color.RGBA{50, 50, 100, 240},
		true,
	)

	// Draw title
	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate(20, 30)
	titleOp.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, "Creature Management", g.fontFace, titleOp)

	if g.menuSection == 0 {
		// Draw creature list
		for i, creature := range g.creatures {
			op := &text.DrawOptions{}
			op.GeoM.Translate(30, float64(60+i*20))

			if i == g.selectedCreature {
				op.ColorScale.ScaleWithColor(color.RGBA{255, 255, 0, 255}) // Yellow for selected

				// Draw selector arrow
				selectorOp := &text.DrawOptions{}
				selectorOp.GeoM.Translate(20, float64(60+i*20))
				selectorOp.ColorScale.ScaleWithColor(color.RGBA{255, 255, 0, 255})
				text.Draw(screen, ">", g.fontFace, selectorOp)
			} else {
				op.ColorScale.ScaleWithColor(color.White)
			}

			// Show creature name and level
			text.Draw(screen, creature.name+" Lv."+strconv.Itoa(creature.level), g.fontFace, op)

			// If this is the active creature, mark it
			if creature.name == g.battle.playerCreature.name {
				activeOp := &text.DrawOptions{}
				activeOp.GeoM.Translate(180, float64(60+i*20))
				activeOp.ColorScale.ScaleWithColor(color.RGBA{0, 255, 0, 255})
				text.Draw(screen, "(Active)", g.fontFace, activeOp)
			}
		}

		// Draw instructions
		instructionsOp := &text.DrawOptions{}
		instructionsOp.GeoM.Translate(20, float64(screenHeight-30))
		instructionsOp.ColorScale.ScaleWithColor(color.RGBA{200, 200, 200, 255})
		text.Draw(screen, "Arrow keys to navigate, Space to select, ESC to exit", g.fontFace, instructionsOp)
	} else if g.menuSection == 1 {
		// Draw creature details
		creature := g.creatures[g.selectedCreature]

		// Draw creature name and type
		nameOp := &text.DrawOptions{}
		nameOp.GeoM.Translate(30, 60)
		nameOp.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, creature.name+" ("+creature.type1+")", g.fontFace, nameOp)

		// Draw HP
		hpOp := &text.DrawOptions{}
		hpOp.GeoM.Translate(30, 80)
		hpOp.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, "HP: "+strconv.Itoa(creature.hp)+"/"+strconv.Itoa(creature.maxHP), g.fontFace, hpOp)

		// Draw stats
		statsOp := &text.DrawOptions{}
		statsOp.GeoM.Translate(30, 100)
		statsOp.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, "Attack: "+strconv.Itoa(creature.attack), g.fontFace, statsOp)

		defOp := &text.DrawOptions{}
		defOp.GeoM.Translate(30, 115)
		defOp.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, "Defense: "+strconv.Itoa(creature.defense), g.fontFace, defOp)

		spdOp := &text.DrawOptions{}
		spdOp.GeoM.Translate(30, 130)
		spdOp.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, "Speed: "+strconv.Itoa(creature.speed), g.fontFace, spdOp)

		// Draw moves
		movesOp := &text.DrawOptions{}
		movesOp.GeoM.Translate(30, 155)
		movesOp.ColorScale.ScaleWithColor(color.White)
		text.Draw(screen, "Moves:", g.fontFace, movesOp)

		for i, move := range creature.moves {
			moveOp := &text.DrawOptions{}
			moveOp.GeoM.Translate(40, float64(175+i*15))
			moveOp.ColorScale.ScaleWithColor(color.White)
			text.Draw(screen, "- "+move.name+" ("+move.type1+")", g.fontFace, moveOp)

			movePowerOp := &text.DrawOptions{}
			movePowerOp.GeoM.Translate(180, float64(175+i*15))
			movePowerOp.ColorScale.ScaleWithColor(color.White)
			text.Draw(screen, "Power: "+strconv.Itoa(move.power), g.fontFace, movePowerOp)
		}

		// Draw menu options
		for i, option := range g.creatureMenuOptions {
			op := &text.DrawOptions{}
			op.GeoM.Translate(float64(screenWidth/2-30), float64(screenHeight-70+i*20))

			if i == g.selectedOption {
				op.ColorScale.ScaleWithColor(color.RGBA{255, 255, 0, 255}) // Yellow for selected

				// Draw selector arrow
				selectorOp := &text.DrawOptions{}
				selectorOp.GeoM.Translate(float64(screenWidth/2-45), float64(screenHeight-70+i*20))
				selectorOp.ColorScale.ScaleWithColor(color.RGBA{255, 255, 0, 255})
				text.Draw(screen, ">", g.fontFace, selectorOp)
			} else {
				op.ColorScale.ScaleWithColor(color.White)
			}

			text.Draw(screen, option, g.fontFace, op)
		}
	}
}
