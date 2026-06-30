package main

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ── Символы барабанов ────────────────────────────────────────────────────
const (
	symCherry = iota
	symLemon
	symBar
	symStar
	symDiamond
	symSeven
	symCount
)

// вес выпадения и множитель за три в ряд
var symWeight = [symCount]int{30, 26, 18, 12, 8, 6}
var symTriple = [symCount]int{5, 8, 10, 15, 25, 50}

func pickSymbol() int {
	total := 0
	for _, w := range symWeight {
		total += w
	}
	r := rand.Intn(total)
	for s, w := range symWeight {
		if r < w {
			return s
		}
		r -= w
	}
	return 0
}

// ── Геометрия автомата ───────────────────────────────────────────────────
const (
	slotCellH = 96
	slotReelW = 130
	slotRows  = 3
	slotGapX  = 16
	slotPad   = 22
)

var (
	reelsW    = float32(3*slotReelW + 2*slotGapX)
	frameW    = reelsW + 2*slotPad
	frameX    = (screenW - frameW) / 2
	frameY    = float32(76)
	reelsTop  = frameY + slotPad
	reel0X    = frameX + slotPad
	windowH   = float32(slotRows * slotCellH)
	paylineY  = reelsTop + windowH/2
	frameH    = windowH + 2*slotPad
)

func reelX(r int) float32 { return reel0X + float32(r)*(slotReelW+slotGapX) }

type sPhase int

const (
	sIdle sPhase = iota
	sSpinning
	sDone
)

type slotsScene struct {
	bet     *BetControl
	btnSpin Button

	strips [3][]int
	winRow [3]int
	result [3]int
	target [3]float64
	dur    [3]float64
	sc     [3]float64

	phase   sPhase
	t       float64
	payout  int
	lastBet int

	cheat bool // зажат пробел во время прокрутки — гарантированный выигрыш
}

func newSlotsScene() *slotsScene {
	s := &slotsScene{}
	// Ряд фишек по ширине автомата.
	s.bet = newBetControl(frameX, frameY+frameH+54, 50, frameW)
	s.btnSpin = Button{
		X: frameX, Y: frameY + frameH + 112, W: frameW, H: 60,
		Label: "КРУТИТЬ", Face: faceBig,
		Bg: color.RGBA{0xe8, 0x5a, 0x5a, 0xff}, Fg: colText,
	}
	s.setupBoard(false)
	return s
}

func (s *slotsScene) Title() string { return "Слоты" }

// setupBoard формирует барабаны. spin=false — сразу показать готовый расклад.
func (s *slotsScene) setupBoard(spin bool) {
	for r := 0; r < 3; r++ {
		s.result[r] = pickSymbol()
		s.winRow[r] = 16 + r*4
		strip := make([]int, s.winRow[r]+3)
		for i := range strip {
			strip[i] = rand.Intn(symCount)
		}
		strip[s.winRow[r]] = s.result[r]
		s.strips[r] = strip
		s.target[r] = float64(s.winRow[r]-1) * slotCellH
		s.dur[r] = 2.4 + float64(r)*0.6
		if spin {
			s.sc[r] = 0
		} else {
			s.sc[r] = s.target[r]
		}
	}
}

func (s *slotsScene) Update(g *Game) error {
	locked := s.phase == sSpinning
	s.bet.Update(g.state.Balance, locked)

	if !locked {
		s.btnSpin.Disabled = s.bet.Amount <= 0 || s.bet.Amount > g.state.Balance
		if s.btnSpin.Clicked() {
			s.spin(g)
		}
	}

	if s.phase == sSpinning {
		// Чит: пока крутятся барабаны и зажат пробел — даём плюс.
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			s.cheat = true
		}
		s.t += 1.0 / 60.0
		maxDur := 0.0
		for r := 0; r < 3; r++ {
			tt := s.t / s.dur[r]
			if tt > 1 {
				tt = 1
			}
			s.sc[r] = s.target[r] * easeOutCubic(tt)
			if s.dur[r] > maxDur {
				maxDur = s.dur[r]
			}
		}
		if s.t >= maxDur {
			s.finish(g)
		}
	}
	return nil
}

func (s *slotsScene) spin(g *Game) {
	s.lastBet = s.bet.Amount
	g.state.Balance -= s.lastBet
	g.state.TotalSpins++
	s.setupBoard(true)
	s.cheat = false
	s.t = 0
	s.phase = sSpinning
}

func (s *slotsScene) finish(g *Game) {
	s.phase = sDone
	if s.cheat {
		// Выстраиваем три одинаковых символа на линии выплаты,
		// синхронно правя видимую ленту, чтобы барабаны показали тройку.
		win := s.result[0]
		for r := 0; r < 3; r++ {
			s.result[r] = win
			s.strips[r][s.winRow[r]] = win
		}
	}
	mult := evalLine(s.result)
	s.payout = s.lastBet * mult
	if s.payout > 0 {
		g.state.Balance += s.payout
		g.state.recordWin(s.payout - s.lastBet)
	}
	g.state.save()
}

// evalLine возвращает множитель выплаты по центральной линии.
func evalLine(line [3]int) int {
	if line[0] == line[1] && line[1] == line[2] {
		return symTriple[line[0]]
	}
	if line[0] == line[1] || line[1] == line[2] || line[0] == line[2] {
		return 2 // любая пара
	}
	return 0
}

func (s *slotsScene) Draw(g *Game, screen *ebiten.Image) {
	// Корпус автомата.
	fillRoundRect(screen, frameX-4, frameY-4, frameW+8, frameH+8, 26, scaleColor(colAccent, 0.45))
	fillRoundRect(screen, frameX, frameY, frameW, frameH, 22, color.RGBA{0x24, 0x1d, 0x33, 0xff})

	// Барабаны.
	for r := 0; r < 3; r++ {
		s.drawReel(screen, r)
	}

	// Линия выплаты поверх барабанов.
	lineClr := color.RGBA{0xe8, 0x5a, 0x5a, 0xb0}
	fillRect(screen, reel0X, paylineY-1.5, reelsW, 3, lineClr)

	s.bet.Draw(screen)
	s.btnSpin.Draw(screen)
	s.drawStatus(screen)
}

func (s *slotsScene) drawReel(screen *ebiten.Image, r int) {
	x := reelX(r)
	// Светлая подложка барабана.
	fillRoundRect(screen, x, reelsTop, slotReelW, windowH, 14, color.RGBA{0xec, 0xe6, 0xda, 0xff})

	clip := screen.SubImage(image.Rect(
		int(x), int(reelsTop), int(x+slotReelW), int(reelsTop+windowH),
	)).(*ebiten.Image)

	cx := x + slotReelW/2
	for i, kind := range s.strips[r] {
		cy := reelsTop + slotCellH/2 + float32(i)*slotCellH - float32(s.sc[r])
		if cy < reelsTop-slotCellH || cy > reelsTop+windowH+slotCellH {
			continue
		}
		drawSymbol(clip, kind, cx, cy)
	}

	// Разделители ячеек и рамка барабана.
	for i := 1; i < slotRows; i++ {
		y := reelsTop + float32(i)*slotCellH
		fillRect(screen, x+8, y, slotReelW-16, 1, color.RGBA{0, 0, 0, 0x22})
	}
	strokeRoundRect(screen, x+1, reelsTop+1, slotReelW-2, windowH-2, 13, 2, scaleColor(colAccent, 0.5))
}

func (s *slotsScene) drawStatus(screen *ebiten.Image) {
	cx := float64(screenW / 2)
	y := float64(frameY + frameH + 196)
	switch s.phase {
	case sIdle:
		drawTextCentered(screen, "Собери три в ряд!", faceMed, cx, y, colTextDim)
	case sSpinning:
		drawTextCentered(screen, "Крутится...", faceMed, cx, y, colText)
	case sDone:
		if s.payout > 0 {
			msg := "Выигрыш +" + coinStr(s.payout-s.lastBet) + "!"
			drawTextCentered(screen, msg, faceMed, cx, y, colAccent)
		} else {
			drawTextCentered(screen, "Мимо: -"+coinStr(s.lastBet), faceMed, cx, y, colLose)
		}
	}
}

// ── Рисование символов ─────────────────────────────────────────────────────
var (
	cherryRed  = color.RGBA{0xd6, 0x2f, 0x2f, 0xff}
	cherryDk   = color.RGBA{0x9c, 0x1f, 0x1f, 0xff}
	lemonY     = color.RGBA{0xf4, 0xd0, 0x3a, 0xff}
	lemonDk    = color.RGBA{0xcf, 0xa8, 0x1f, 0xff}
	diamondCol = color.RGBA{0x4f, 0xd2, 0xe0, 0xff}
	barCol     = color.RGBA{0x7a, 0x4b, 0xe0, 0xff}
	highlight  = color.RGBA{0xff, 0xff, 0xff, 0xcc}
)

func drawSymbol(dst *ebiten.Image, kind int, cx, cy float32) {
	switch kind {
	case symCherry:
		stem := &vector.Path{}
		stem.MoveTo(cx-14, cy+10)
		stem.QuadTo(cx+2, cy-30, cx+6, cy-34)
		stem.MoveTo(cx+16, cy+14)
		stem.QuadTo(cx+12, cy-26, cx+6, cy-34)
		strokePath(dst, stem, 4, colGreen)
		fillPath(dst, roundRectPath(cx+4, cy-42, 24, 13, 6), colGreen)
		fillCircle(dst, cx-14, cy+18, 16, cherryDk)
		fillCircle(dst, cx-14, cy+18, 13, cherryRed)
		fillCircle(dst, cx+16, cy+20, 16, cherryDk)
		fillCircle(dst, cx+16, cy+20, 13, cherryRed)
		fillCircle(dst, cx-18, cy+13, 4, highlight)
	case symLemon:
		fillRoundRect(dst, cx-32, cy-19, 64, 42, 21, lemonDk)
		fillRoundRect(dst, cx-30, cy-17, 60, 38, 19, lemonY)
		fillPath(dst, roundRectPath(cx-4, cy-30, 22, 11, 5), colGreen)
		fillCircle(dst, cx-14, cy-6, 4, highlight)
	case symBar:
		fillRoundRect(dst, cx-34, cy-18, 68, 36, 9, barCol)
		fillRoundRect(dst, cx-30, cy-14, 60, 28, 7, scaleColor(barCol, 1.25))
		drawTextCentered(dst, "BAR", faceMed, float64(cx), float64(cy), colText)
	case symStar:
		fillPath(dst, starPath(cx, cy-2, 36, 16), scaleColor(colAccent, 0.7))
		fillPath(dst, starPath(cx, cy-2, 31, 14), colAccent)
	case symDiamond:
		g := &vector.Path{}
		g.MoveTo(cx, cy-28)
		g.LineTo(cx+28, cy-6)
		g.LineTo(cx, cy+30)
		g.LineTo(cx-28, cy-6)
		g.Close()
		fillPath(dst, g, diamondCol)
		fl := &vector.Path{}
		fl.MoveTo(cx-28, cy-6)
		fl.LineTo(cx+28, cy-6)
		fl.MoveTo(cx-14, cy-17)
		fl.LineTo(cx, cy-6)
		fl.LineTo(cx+14, cy-17)
		strokePath(dst, fl, 2, scaleColor(diamondCol, 1.4))
	case symSeven:
		drawTextCentered(dst, "7", faceHuge, float64(cx+2), float64(cy+2), cherryDk)
		drawTextCentered(dst, "7", faceHuge, float64(cx), float64(cy), colRed)
	}
}

// starPath строит путь пятиконечной звезды.
func starPath(cx, cy, outer, inner float32) *vector.Path {
	p := &vector.Path{}
	for i := 0; i < 10; i++ {
		r := outer
		if i%2 == 1 {
			r = inner
		}
		a := -math.Pi/2 + float64(i)*math.Pi/5
		x := cx + r*float32(math.Cos(a))
		y := cy + r*float32(math.Sin(a))
		if i == 0 {
			p.MoveTo(x, y)
		} else {
			p.LineTo(x, y)
		}
	}
	p.Close()
	return p
}
