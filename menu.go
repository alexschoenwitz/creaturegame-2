package main

import (
	"image/color"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// updateMainMenu handles main menu state updates
func (g *Game) updateMainMenu() {
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.selectedOption = (g.selectedOption - 1 + len(g.menuOptions)) % len(g.menuOptions)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.selectedOption = (g.selectedOption + 1) % len(g.menuOptions)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		switch g.selectedOption {
		case 0: // New Game
			g.initGame()
			g.gameState = StateOverworld
		case 1: // Options - could be implemented later
			// For now, just print to console
			log.Println("Options selected (not implemented)")
		case 2: // Exit
			os.Exit(0)
			// return errors.New("exit game")
		}
	}
}

// drawMainMenu draws the main menu
func (g *Game) drawMainMenu(screen *ebiten.Image) {
	// Draw title
	titleOp := &text.DrawOptions{}
	titleOp.GeoM.Translate(float64(screenWidth/2-50), float64(screenHeight/4))
	titleOp.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, 255})
	text.Draw(screen, "CreatureGame", g.fontFace, titleOp)

	// Draw menu options
	for i, option := range g.menuOptions {
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(screenWidth/2-30), float64(screenHeight/2+i*20))

		// Highlight selected option
		if i == g.selectedOption {
			op.ColorScale.ScaleWithColor(color.RGBA{255, 255, 0, 255}) // Yellow for selected

			// Draw selector arrow
			selectorOp := &text.DrawOptions{}
			selectorOp.GeoM.Translate(float64(screenWidth/2-45), float64(screenHeight/2+i*20))
			selectorOp.ColorScale.ScaleWithColor(color.RGBA{255, 255, 0, 255})
			text.Draw(screen, ">", g.fontFace, selectorOp)
		} else {
			op.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, 255}) // White for unselected
		}

		text.Draw(screen, option, g.fontFace, op)
	}

	// Draw instructions
	instructionsOp := &text.DrawOptions{}
	instructionsOp.GeoM.Translate(10, float64(screenHeight-25))
	instructionsOp.ColorScale.ScaleWithColor(color.RGBA{200, 200, 200, 255})
	text.Draw(screen, "Arrow keys to navigate, Space/Enter to select", g.fontFace, instructionsOp)
}
