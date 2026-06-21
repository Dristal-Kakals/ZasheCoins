package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenW = 980
	screenH = 620
	topBarH = 64
)

// Scene — экран игры.
type Scene interface {
	Update(g *Game) error
	Draw(g *Game, screen *ebiten.Image)
	Title() string
}

// Game — корневая структура Ebiten.
type Game struct {
	state *State
	scene Scene
	// общие кнопки верхней панели
	btnMenu Button
}

func NewGame() *Game {
	g := &Game{state: loadState()}
	g.btnMenu = Button{W: 120, H: 40, Label: "‹ Меню", Face: faceMed}
	g.scene = newMenuScene()
	return g
}

func (g *Game) goTo(s Scene) {
	g.scene = s
}

func (g *Game) Update() error {
	// Кнопка возврата в меню (везде, кроме самого меню).
	if _, isMenu := g.scene.(*menuScene); !isMenu {
		g.btnMenu.X, g.btnMenu.Y = 16, topBarH/2-20
		if g.btnMenu.Clicked() {
			g.state.save()
			g.goTo(newMenuScene())
			return nil
		}
	}
	return g.scene.Update(g)
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(colBg)
	g.scene.Draw(g, screen)
	g.drawTopBar(screen)
}

func (g *Game) drawTopBar(screen *ebiten.Image) {
	fillRect(screen, 0, 0, screenW, topBarH, colPanel)
	fillRect(screen, 0, topBarH-2, screenW, 2, colPanel2)

	if _, isMenu := g.scene.(*menuScene); !isMenu {
		g.btnMenu.Draw(screen)
	}

	// Заголовок текущей сцены по центру.
	drawTextCentered(screen, g.scene.Title(), faceMed, screenW/2, topBarH/2, colTextDim)

	// Баланс справа.
	label := coinStr(g.state.Balance)
	tw := textWidth(label, faceBig)
	drawText(screen, label, faceBig, screenW-tw-24, topBarH/2-22, colAccent)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenW, screenH
}

func main() {
	initFonts()
	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("ZasheCoins казик")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
