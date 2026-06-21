package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type menuScene struct {
	btnRoulette Button
	btnSlots    Button
	btnCases    Button
	btnDeposit  Button
	t           float64
}

func newMenuScene() *menuScene {
	m := &menuScene{}
	cx := float32(screenW/2 - 170)
	m.btnRoulette = Button{X: cx, Y: 284, W: 340, H: 62, Label: "Рулетка", Face: faceBig}
	m.btnSlots = Button{X: cx, Y: 354, W: 340, H: 62, Label: "Слоты", Face: faceBig}
	m.btnCases = Button{X: cx, Y: 424, W: 340, H: 62, Label: "Кейсы", Face: faceBig}
	m.btnDeposit = Button{X: float32(screenW/2 - 130), Y: 502, W: 260, H: 44,
		Label: "Внести додеп +" + formatCoins(depositAmount), Face: faceSmall, Bg: colGreen}
	return m
}

func (m *menuScene) Title() string { return "" }

func (m *menuScene) Update(g *Game) error {
	m.t += 1.0 / 60.0
	if m.btnRoulette.Clicked() {
		g.goTo(newRouletteScene())
	}
	if m.btnSlots.Clicked() {
		g.goTo(newSlotsScene())
	}
	if m.btnCases.Clicked() {
		g.goTo(newCasesScene())
	}
	// Додеп доступен только при балансе не выше порога.
	m.btnDeposit.Disabled = g.state.Balance >= depositMaxBalance
	if m.btnDeposit.Clicked() {
		g.state.Balance += depositAmount
		g.state.save()
	}
	return nil
}

func (m *menuScene) Draw(g *Game, screen *ebiten.Image) {
	cx := float32(screenW / 2)

	// Декоративная золотая монета над названием.
	bob := float32(7 * math.Sin(m.t*2))
	drawCoin(screen, cx, 118+bob, 36)

	// Название.
	drawTextCentered(screen, "ZasheCoins", faceHuge, screenW/2, 196, colAccent)
	drawTextCentered(screen, "к а з и к", faceMed, screenW/2, 238, colText)
	// Акцентный разделитель.
	fillRoundRect(screen, cx-90, 260, 180, 3, 1.5, colPanel2)

	m.btnRoulette.Draw(screen)
	m.btnSlots.Draw(screen)
	m.btnCases.Draw(screen)
	m.btnDeposit.Draw(screen)
	if m.btnDeposit.Disabled {
		hint := fmt.Sprintf("Додеп доступен при балансе ниже %s", coinStr(depositMaxBalance))
		drawTextCentered(screen, hint, faceSmall, screenW/2, 558, colTextDim)
	}

	// Статистика снизу.
	stats := fmt.Sprintf("Прокруток: %d     Кейсов открыто: %d     Лучший выигрыш: %s",
		g.state.TotalSpins, g.state.CasesOpened, coinStr(g.state.BiggestWin))
	drawTextCentered(screen, stats, faceSmall, screenW/2, 594, colTextDim)
}

// drawCoin рисует стилизованную монету защекоина с буквой Z.
func drawCoin(dst *ebiten.Image, cx, cy, r float32) {
	fillCircle(dst, cx, cy+3, r, color.RGBA{0, 0, 0, 0x55}) // тень
	fillCircle(dst, cx, cy, r, scaleColor(colAccent, 0.75)) // ободок
	fillCircle(dst, cx, cy, r-5, colAccent)                 // тело
	fillCircle(dst, cx, cy, r-9, scaleColor(colAccent, 1.12))
	drawTextCentered(dst, "Z", faceBig, float64(cx), float64(cy), colBg)
}
