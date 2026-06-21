package main

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const slots = 37 // 0..36, европейская рулетка

// Красные номера европейской рулетки.
var redNumbers = map[int]bool{
	1: true, 3: true, 5: true, 7: true, 9: true, 12: true, 14: true, 16: true,
	18: true, 19: true, 21: true, 23: true, 25: true, 27: true, 30: true, 32: true, 34: true, 36: true,
}

func slotColor(n int) color.RGBA {
	switch {
	case n == 0:
		return colGreen
	case redNumbers[n]:
		return colRed
	default:
		return colBlack
	}
}

type rouletteBet int

const (
	betRed rouletteBet = iota
	betBlack
	betGreen
)

type rPhase int

const (
	rIdle rPhase = iota
	rSpinning
	rDone
)

type rouletteScene struct {
	bet      *BetControl
	choice   rouletteBet
	btnRed   Button
	btnBlack Button
	btnGreen Button
	btnSpin  Button

	wheel    *ebiten.Image
	wheelR   float32
	wcx, wcy float32

	phase     rPhase
	theta     float64 // текущий угол поворота
	startθ    float64
	targetθ   float64
	spinT     float64
	spinDur   float64
	result    int
	lastBet   int
	lastColor rouletteBet
	payout    int
}

func newRouletteScene() *rouletteScene {
	s := &rouletteScene{}
	s.choice = betRed
	// Колесо — слева, вся панель управления — в правой колонке (512..941).
	s.wcx, s.wcy = 300, 360
	s.wheelR = 170
	s.wheel = buildWheelImage(s.wheelR)

	s.bet = newBetControl(512, 270, 100, 429)
	by := float32(360)
	s.btnRed = Button{X: 512, Y: by, W: 135, H: 60, Label: "Красное x2", Face: faceSmall, Bg: colRed}
	s.btnBlack = Button{X: 659, Y: by, W: 135, H: 60, Label: "Чёрное x2", Face: faceSmall, Bg: colBlack}
	s.btnGreen = Button{X: 806, Y: by, W: 135, H: 60, Label: "Зеро x14", Face: faceSmall, Bg: colGreen}
	s.btnSpin = Button{X: 512, Y: 446, W: 429, H: 74, Label: "КРУТИТЬ", Face: faceBig, Bg: colAccent, Fg: colBg}
	return s
}

func (s *rouletteScene) Title() string { return "Рулетка" }

func (s *rouletteScene) Update(g *Game) error {
	locked := s.phase == rSpinning
	s.bet.Update(g.state.Balance, locked)

	if !locked {
		if s.btnRed.Clicked() {
			s.choice = betRed
		}
		if s.btnBlack.Clicked() {
			s.choice = betBlack
		}
		if s.btnGreen.Clicked() {
			s.choice = betGreen
		}
		s.btnSpin.Disabled = s.bet.Amount <= 0 || s.bet.Amount > g.state.Balance
		if s.btnSpin.Clicked() {
			s.startSpin(g)
		}
	}

	if s.phase == rSpinning {
		s.spinT += 1.0 / 60.0
		t := s.spinT / s.spinDur
		if t >= 1 {
			t = 1
			s.finish(g)
		}
		s.theta = s.startθ + (s.targetθ-s.startθ)*easeOutCubic(t)
	}
	return nil
}

func (s *rouletteScene) startSpin(g *Game) {
	s.lastBet = s.bet.Amount
	s.lastColor = s.choice
	g.state.Balance -= s.bet.Amount
	g.state.TotalSpins++

	s.result = rand.Intn(slots)
	seg := 2 * math.Pi / float64(slots)
	mid := (float64(s.result) + 0.5) * seg
	// θ так, чтобы сектор result оказался под указателем (вверху, -π/2).
	base := -math.Pi/2 - mid
	// нормализуем относительно текущего и добавляем полные обороты
	s.startθ = s.theta
	for base < s.startθ {
		base += 2 * math.Pi
	}
	s.targetθ = base + 2*math.Pi*float64(5+rand.Intn(3))
	s.spinT = 0
	s.spinDur = 4.5
	s.phase = rSpinning
}

func (s *rouletteScene) finish(g *Game) {
	s.phase = rDone
	resCol := slotColor(s.result)
	mult := 0
	switch s.lastColor {
	case betRed:
		if resCol == colRed {
			mult = 2
		}
	case betBlack:
		if resCol == colBlack {
			mult = 2
		}
	case betGreen:
		if s.result == 0 {
			mult = 14
		}
	}
	s.payout = s.lastBet * mult
	if s.payout > 0 {
		g.state.Balance += s.payout
		net := s.payout - s.lastBet
		g.state.recordWin(net)
	}
	g.state.save()
}

func (s *rouletteScene) Draw(g *Game, screen *ebiten.Image) {
	// Колесо.
	op := &ebiten.DrawImageOptions{}
	iw, ih := s.wheel.Bounds().Dx(), s.wheel.Bounds().Dy()
	op.GeoM.Translate(-float64(iw)/2, -float64(ih)/2)
	op.GeoM.Rotate(s.theta)
	op.GeoM.Translate(float64(s.wcx), float64(s.wcy))
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(s.wheel, op)

	// Ступица + указатель.
	drawCoin(screen, s.wcx, s.wcy, 40)
	drawPointer(screen, s.wcx, s.wcy-s.wheelR-6)

	// Панель ставок справа.
	s.btnRed.Active = s.choice == betRed
	s.btnBlack.Active = s.choice == betBlack
	s.btnGreen.Active = s.choice == betGreen
	drawText(screen, "На что ставим:", faceSmall, 512, 338, colTextDim)
	s.btnRed.Draw(screen)
	s.btnBlack.Draw(screen)
	s.btnGreen.Draw(screen)
	s.btnSpin.Draw(screen)

	s.bet.Draw(screen)

	// Результат / статус.
	s.drawStatus(screen)
}

func (s *rouletteScene) drawStatus(screen *ebiten.Image) {
	x := float64(514)
	switch s.phase {
	case rIdle:
		drawText(screen, "Сделай ставку,", faceMed, x, 120, colText)
		drawText(screen, "выбери цвет и крути!", faceMed, x, 152, colTextDim)
	case rSpinning:
		drawText(screen, "Колесо крутится...", faceMed, x, 136, colText)
	case rDone:
		col := slotColor(s.result)
		drawText(screen, "Выпало число:", faceMed, x, 110, colTextDim)
		cy := float32(176)
		fillCircle(screen, float32(x)+34, cy, 30, scaleColor(col, 0.6))
		fillCircle(screen, float32(x)+34, cy, 26, col)
		drawTextCentered(screen, intStr(s.result), faceBig, x+34, float64(cy), colText)
		if s.payout > 0 {
			drawText(screen, "Выигрыш +"+coinStr(s.payout-s.lastBet), faceMed, x+86, 162, colWin)
		} else {
			drawText(screen, "Мимо: -"+coinStr(s.lastBet), faceMed, x+86, 162, colLose)
		}
	}
}

func intStr(n int) string { return formatCoins(n) }

// drawPointer рисует треугольник-указатель, направленный вниз.
func drawPointer(dst *ebiten.Image, x, y float32) {
	var p vector.Path
	p.MoveTo(x-16, y-4)
	p.LineTo(x+16, y-4)
	p.LineTo(x, y+22)
	p.Close()
	fillPath(dst, &p, colAccent)
}

// buildWheelImage заранее рендерит колесо рулетки.
func buildWheelImage(radius float32) *ebiten.Image {
	pad := float32(8)
	size := int(2*(radius+pad)) + 2
	img := ebiten.NewImage(size, size)
	cx, cy := float32(size)/2, float32(size)/2
	seg := 2 * math.Pi / float64(slots)

	// Внешнее кольцо.
	fillCircle(img, cx, cy, radius+pad, color.RGBA{0x3a, 0x2c, 0x12, 0xff})

	for n := 0; n < slots; n++ {
		a0 := float64(n) * seg
		a1 := float64(n+1) * seg
		var p vector.Path
		p.MoveTo(cx, cy)
		p.Arc(cx, cy, radius, float32(a0), float32(a1), vector.Clockwise)
		p.Close()
		fillPath(img, &p, slotColor(n))

		// Номер сектора.
		mid := (a0 + a1) / 2
		r2 := float64(radius) * 0.82
		nx := float64(cx) + r2*math.Cos(mid)
		ny := float64(cy) + r2*math.Sin(mid)
		drawTextCentered(img, intStr(n), faceSmall, nx, ny, colText)
	}
	return img
}
