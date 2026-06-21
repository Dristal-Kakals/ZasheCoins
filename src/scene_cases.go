package main

import (
	"image"
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Item — приз в кейсе.
type Item struct {
	Name   string
	Value  int
	Weight int
	Rarity color.RGBA
}

// Case — кейс с ценой и набором призов.
type Case struct {
	Name  string
	Price int
	Items []Item
}

var cases = []Case{
	{
		Name: "Деревянный", Price: 100,
		Items: []Item{
			{"Ржавый болт", 10, 40, rarCommon},
			{"Медяк", 50, 30, rarCommon},
			{"Серебрушка", 120, 18, rarUncommon},
			{"Самоцвет", 300, 9, rarRare},
			{"Золотой слиток", 800, 3, rarEpic},
			{"Корона", 2500, 1, rarLegendary},
		},
	},
	{
		Name: "Стальной", Price: 500,
		Items: []Item{
			{"Гайка", 80, 38, rarCommon},
			{"Монетница", 300, 30, rarCommon},
			{"Кубок", 700, 18, rarUncommon},
			{"Рубин", 1500, 9, rarRare},
			{"Сейф", 4000, 3, rarEpic},
			{"Джекпот", 12000, 1, rarLegendary},
		},
	},
	{
		Name: "Алмазный", Price: 2000,
		Items: []Item{
			{"Осколок", 400, 38, rarCommon},
			{"Чемодан", 1500, 30, rarCommon},
			{"Изумруд", 3500, 18, rarUncommon},
			{"Бриллиант", 7000, 9, rarRare},
			{"Слиток платины", 18000, 3, rarEpic},
			{"МЕГА-ДЖЕКПОТ", 60000, 1, rarLegendary},
		},
	},
}

const (
	tileW = 120
	tileH = 150
	gap   = 12
	pitch = tileW + gap
)

type cPhase int

const (
	cIdle cPhase = iota
	cSpinning
	cDone
)

type casesScene struct {
	sel       int // выбранный кейс
	caseBtns  []Button
	btnOpen   Button

	phase  cPhase
	reel   []int // индексы предметов в ленте
	winIdx int
	won    Item
	scroll float64
	target float64
	t      float64
	dur    float64
}

func newCasesScene() *casesScene {
	s := &casesScene{}
	x := float32(40)
	for range cases {
		s.caseBtns = append(s.caseBtns, Button{
			X: x, Y: 352, W: 200, H: 78, Face: faceMed, Bg: colPanel2,
		})
		x += 216
	}
	s.btnOpen = Button{X: 40, Y: 470, W: 632, H: 70, Label: "ОТКРЫТЬ", Face: faceBig, Bg: colAccent, Fg: colBg}
	return s
}

func (s *casesScene) Title() string { return "Кейсы" }

func (s *casesScene) cur() *Case { return &cases[s.sel] }

func (s *casesScene) Update(g *Game) error {
	locked := s.phase == cSpinning
	if !locked {
		for i := range s.caseBtns {
			if s.caseBtns[i].Clicked() {
				s.sel = i
				s.phase = cIdle
			}
		}
		price := s.cur().Price
		s.btnOpen.Label = "ОТКРЫТЬ за " + coinStr(price)
		s.btnOpen.Disabled = g.state.Balance < price
		if s.btnOpen.Clicked() {
			s.open(g)
		}
	}

	if s.phase == cSpinning {
		s.t += 1.0 / 60.0
		tt := s.t / s.dur
		if tt >= 1 {
			tt = 1
			s.finish(g)
		}
		s.scroll = s.target * easeOutCubic(tt)
	}
	return nil
}

func (s *casesScene) open(g *Game) {
	c := s.cur()
	g.state.Balance -= c.Price
	g.state.CasesOpened++

	s.won = pickItem(c)
	// Строим ленту: победитель ближе к концу.
	s.winIdx = 58 + rand.Intn(6)
	s.reel = make([]int, s.winIdx+12)
	for i := range s.reel {
		s.reel[i] = rand.Intn(len(c.Items))
	}
	// найдём индекс выигрышного предмета в массиве предметов
	wi := 0
	for i := range c.Items {
		if c.Items[i].Name == s.won.Name {
			wi = i
			break
		}
	}
	s.reel[s.winIdx] = wi

	jitter := float64(rand.Intn(tileW-40) - (tileW-40)/2)
	s.target = float64(s.winIdx)*pitch + tileW/2 + jitter
	s.t = 0
	s.dur = 5.0
	s.scroll = 0
	s.phase = cSpinning
}

func (s *casesScene) finish(g *Game) {
	s.phase = cDone
	g.state.Balance += s.won.Value
	net := s.won.Value - s.cur().Price
	if net > 0 {
		g.state.recordWin(net)
	}
	g.state.save()
}

func pickItem(c *Case) Item {
	total := 0
	for _, it := range c.Items {
		total += it.Weight
	}
	r := rand.Intn(total)
	for _, it := range c.Items {
		if r < it.Weight {
			return it
		}
		r -= it.Weight
	}
	return c.Items[len(c.Items)-1]
}

const (
	reelY = float32(116)
	panelX = float32(24)
	panelW = float32(screenW - 48)
	panelY = reelY - 18
	panelH = float32(tileH + 36)
)

func (s *casesScene) Draw(g *Game, screen *ebiten.Image) {
	centerX := float32(screenW / 2)

	// Фон-панель ленты.
	fillRoundRect(screen, panelX, panelY, panelW, panelH, 16, colPanel)
	fillRoundRect(screen, panelX+8, reelY-8, panelW-16, tileH+16, 12, scaleColor(colBg, 1.1))

	// Лента предметов с обрезкой по внутренней области панели.
	clipRect := image.Rect(int(panelX+8), int(reelY-8), int(panelX+panelW-8), int(reelY+tileH+8))
	clip := screen.SubImage(clipRect).(*ebiten.Image)
	if s.phase == cSpinning || s.phase == cDone {
		s.drawReel(clip, centerX, reelY)
	} else {
		s.drawPreview(clip, reelY)
	}

	// Затухание по краям ленты.
	drawEdgeFade(screen, panelX+8, reelY-8, tileH+16)

	// Центральный маркер с указателями.
	drawMarker(screen, centerX, reelY, tileH)

	// Заголовок выбора.
	drawText(screen, "ВЫБЕРИ КЕЙС", faceSmall, float64(panelX)+16, 326, colTextDim)

	// Кнопки кейсов (название + цена внутри, без наложения).
	for i := range s.caseBtns {
		b := &s.caseBtns[i]
		b.Active = s.sel == i
		b.Draw(screen)
		nameClr := colText
		priceClr := colAccent
		if b.Active {
			nameClr, priceClr = colBg, colBg
		}
		cx := float64(b.X + b.W/2)
		drawTextCentered(screen, cases[i].Name, faceMed, cx, float64(b.Y)+26, nameClr)
		drawTextCentered(screen, coinStr(cases[i].Price), faceSmall, cx, float64(b.Y)+54, priceClr)
	}

	s.btnOpen.Draw(screen)
	s.drawInfo(screen)
}

func (s *casesScene) drawReel(dst *ebiten.Image, centerX, reelY float32) {
	c := s.cur()
	for i, idx := range s.reel {
		x := centerX - float32(s.scroll) + float32(i)*pitch
		if x+tileW < 0 || x > screenW {
			continue
		}
		win := s.phase == cDone && i == s.winIdx
		drawTile(dst, x, reelY, c.Items[idx], win)
	}
}

// drawPreview показывает содержимое кейса до открытия (по центру панели).
func (s *casesScene) drawPreview(dst *ebiten.Image, reelY float32) {
	c := s.cur()
	total := float32(len(c.Items))*tileW + float32(len(c.Items)-1)*gap
	x := (screenW - total) / 2
	for _, it := range c.Items {
		drawTile(dst, x, reelY, it, false)
		x += pitch
	}
}

func drawTile(dst *ebiten.Image, x, y float32, it Item, win bool) {
	r := float32(12)
	// тень
	fillRoundRect(dst, x+2, y+3, tileW, tileH, r, color.RGBA{0, 0, 0, 0x55})
	// подложка с лёгким оттенком редкости сверху
	fillRoundRect(dst, x, y, tileW, tileH, r, colPanel2)
	fillRoundRect(dst, x+3, y+3, tileW-6, tileH*0.45, r-2, scaleColor(it.Rarity, 0.35))
	fillRoundRect(dst, x+3, y+tileH*0.30, tileW-6, tileH*0.6, r-2, colPanel2)

	// иконка-кольцо
	cx := x + tileW/2
	icy := y + 56
	fillCircle(dst, cx, icy, 30, scaleColor(it.Rarity, 0.55))
	fillCircle(dst, cx, icy, 26, it.Rarity)
	fillCircle(dst, cx, icy, 18, colPanel2)

	drawTextCentered(dst, it.Name, faceSmall, float64(cx), float64(y+106), colText)

	// нижняя плашка с ценой
	fillRoundRect(dst, x+8, y+tileH-30, tileW-16, 23, 8, it.Rarity)
	drawTextCentered(dst, coinShort(it.Value), faceSmall, float64(cx), float64(y+tileH-18), colBg)

	// рамка редкости (толще для выигрыша)
	bw := float32(2)
	if win {
		bw = 4
	}
	strokeRoundRect(dst, x+1, y+1, tileW-2, tileH-2, r-1, bw, it.Rarity)
}

// drawMarker рисует центральную линию и треугольники-указатели.
func drawMarker(dst *ebiten.Image, cx, top, h float32) {
	fillRect(dst, cx-1.5, top-6, 3, h+12, colAccent)
	var up, down vector.Path
	down.MoveTo(cx-9, top-14)
	down.LineTo(cx+9, top-14)
	down.LineTo(cx, top-2)
	down.Close()
	up.MoveTo(cx-9, top+h+14)
	up.LineTo(cx+9, top+h+14)
	up.LineTo(cx, top+h+2)
	up.Close()
	fillPath(dst, &down, colAccent)
	fillPath(dst, &up, colAccent)
}

// drawEdgeFade рисует затемнение у левого и правого краёв ленты.
func drawEdgeFade(dst *ebiten.Image, x, y, h float32) {
	const fw = 60
	steps := 24
	for i := 0; i < steps; i++ {
		a := uint8(float64(0xe0) * (1 - float64(i)/float64(steps)))
		sw := fw / float32(steps)
		c := color.RGBA{colPanel.R, colPanel.G, colPanel.B, a}
		fillRect(dst, x+float32(i)*sw, y, sw+1, h, c)
		fillRect(dst, x+panelW-16-float32(i)*sw-sw, y, sw+1, h, c)
	}
}

func (s *casesScene) drawInfo(screen *ebiten.Image) {
	px, py := float32(700), float32(346)
	pw, ph := float32(screenW)-px-24, float32(194)
	fillRoundRect(screen, px, py, pw, ph, 14, colPanel)
	strokeRoundRect(screen, px+1, py+1, pw-2, ph-2, 13, 1.5, colPanel2)

	x := float64(px) + 18
	drawText(screen, s.cur().Name, faceMed, x, float64(py)+14, colAccent)
	switch s.phase {
	case cIdle:
		drawText(screen, "Выбери кейс", faceSmall, x, float64(py)+58, colTextDim)
		drawText(screen, "и открывай!", faceSmall, x, float64(py)+80, colTextDim)
	case cSpinning:
		drawText(screen, "Открываем...", faceMed, x, float64(py)+70, colText)
	case cDone:
		drawText(screen, "Выпало:", faceSmall, x, float64(py)+54, colTextDim)
		drawText(screen, s.won.Name, faceMed, x, float64(py)+78, s.won.Rarity)
		drawText(screen, coinStr(s.won.Value), faceMed, x, float64(py)+112, colAccent)
		net := s.won.Value - s.cur().Price
		if net >= 0 {
			drawText(screen, "Профит +"+coinStr(net), faceSmall, x, float64(py)+150, colWin)
		} else {
			drawText(screen, "Минус "+coinStr(-net), faceSmall, x, float64(py)+150, colLose)
		}
	}
}
