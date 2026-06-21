package main

import (
	"bytes"
	"fmt"
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"
)

// ── Палитра ──────────────────────────────────────────────────────────────
var (
	colBg       = color.RGBA{0x12, 0x14, 0x1c, 0xff}
	colPanel    = color.RGBA{0x1e, 0x22, 0x30, 0xff}
	colPanel2   = color.RGBA{0x2a, 0x30, 0x42, 0xff}
	colAccent   = color.RGBA{0xf5, 0xc1, 0x42, 0xff} // золото — защекоины
	colText     = color.RGBA{0xe8, 0xea, 0xf2, 0xff}
	colTextDim  = color.RGBA{0x8a, 0x90, 0xa6, 0xff}
	colRed       = color.RGBA{0xd6, 0x3a, 0x3a, 0xff}
	colBlack     = color.RGBA{0x23, 0x27, 0x34, 0xff}
	colGreen     = color.RGBA{0x2e, 0xa0, 0x55, 0xff}
	colWin       = color.RGBA{0x4c, 0xd1, 0x6e, 0xff}
	colLose      = color.RGBA{0xe0, 0x55, 0x55, 0xff}
	colBtn       = color.RGBA{0x33, 0x3b, 0x52, 0xff}
	colBtnHover  = color.RGBA{0x44, 0x4f, 0x6e, 0xff}
	colBtnActive = color.RGBA{0xf5, 0xc1, 0x42, 0xff}
)

// Редкости для кейсов
var (
	rarCommon    = color.RGBA{0x9a, 0xa3, 0xbd, 0xff}
	rarUncommon  = color.RGBA{0x4b, 0x8d, 0xff, 0xff}
	rarRare      = color.RGBA{0x8a, 0x4b, 0xff, 0xff}
	rarEpic      = color.RGBA{0xd0, 0x4b, 0xe0, 0xff}
	rarLegendary = color.RGBA{0xf5, 0xa6, 0x23, 0xff}
)

// ── Шрифты ───────────────────────────────────────────────────────────────
var (
	faceSmall *text.GoTextFace
	faceMed   *text.GoTextFace
	faceBig   *text.GoTextFace
	faceHuge  *text.GoTextFace
)

func initFonts() {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(err)
	}
	faceSmall = &text.GoTextFace{Source: src, Size: 16}
	faceMed = &text.GoTextFace{Source: src, Size: 22}
	faceBig = &text.GoTextFace{Source: src, Size: 34}
	faceHuge = &text.GoTextFace{Source: src, Size: 58}
}

// whiteImage — однопиксельная белая текстура для заливки путей.
var whiteImage = ebiten.NewImage(3, 3)

func init() {
	whiteImage.Fill(color.White)
}

// ── Текст ────────────────────────────────────────────────────────────────
func drawText(dst *ebiten.Image, s string, face text.Face, x, y float64, clr color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(dst, s, face, op)
}

func drawTextCentered(dst *ebiten.Image, s string, face text.Face, cx, cy float64, clr color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(cx, cy)
	op.ColorScale.ScaleWithColor(clr)
	op.LayoutOptions.PrimaryAlign = text.AlignCenter
	op.LayoutOptions.SecondaryAlign = text.AlignCenter
	text.Draw(dst, s, face, op)
}

// ── Примитивы ────────────────────────────────────────────────────────────
func fillRect(dst *ebiten.Image, x, y, w, h float32, clr color.Color) {
	vector.DrawFilledRect(dst, x, y, w, h, clr, true)
}

func strokeRect(dst *ebiten.Image, x, y, w, h, sw float32, clr color.Color) {
	vector.StrokeRect(dst, x, y, w, h, sw, clr, true)
}

func fillCircle(dst *ebiten.Image, cx, cy, r float32, clr color.Color) {
	vector.DrawFilledCircle(dst, cx, cy, r, clr, true)
}

func colorFloats(clr color.Color) (r, g, b, a float32) {
	cr, cg, cb, ca := clr.RGBA()
	return float32(cr) / 0xffff, float32(cg) / 0xffff, float32(cb) / 0xffff, float32(ca) / 0xffff
}

// fillPath заливает векторный путь сплошным цветом.
func fillPath(dst *ebiten.Image, path *vector.Path, clr color.Color) {
	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	fr, fg, fb, fa := colorFloats(clr)
	for i := range vs {
		vs[i].SrcX, vs[i].SrcY = 1, 1
		vs[i].ColorR, vs[i].ColorG, vs[i].ColorB, vs[i].ColorA = fr, fg, fb, fa
	}
	op := &ebiten.DrawTrianglesOptions{AntiAlias: true}
	dst.DrawTriangles(vs, is, whiteImage, op)
}

// strokePath обводит путь линией заданной толщины.
func strokePath(dst *ebiten.Image, path *vector.Path, sw float32, clr color.Color) {
	op := &vector.StrokeOptions{Width: sw, LineJoin: vector.LineJoinRound, LineCap: vector.LineCapRound}
	vs, is := path.AppendVerticesAndIndicesForStroke(nil, nil, op)
	fr, fg, fb, fa := colorFloats(clr)
	for i := range vs {
		vs[i].SrcX, vs[i].SrcY = 1, 1
		vs[i].ColorR, vs[i].ColorG, vs[i].ColorB, vs[i].ColorA = fr, fg, fb, fa
	}
	dst.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{AntiAlias: true})
}

// roundRectPath строит путь скруглённого прямоугольника.
func roundRectPath(x, y, w, h, r float32) *vector.Path {
	if r > w/2 {
		r = w / 2
	}
	if r > h/2 {
		r = h / 2
	}
	p := &vector.Path{}
	p.MoveTo(x+r, y)
	p.LineTo(x+w-r, y)
	p.ArcTo(x+w, y, x+w, y+r, r)
	p.LineTo(x+w, y+h-r)
	p.ArcTo(x+w, y+h, x+w-r, y+h, r)
	p.LineTo(x+r, y+h)
	p.ArcTo(x, y+h, x, y+h-r, r)
	p.LineTo(x, y+r)
	p.ArcTo(x, y, x+r, y, r)
	p.Close()
	return p
}

func fillRoundRect(dst *ebiten.Image, x, y, w, h, r float32, clr color.Color) {
	fillPath(dst, roundRectPath(x, y, w, h, r), clr)
}

func strokeRoundRect(dst *ebiten.Image, x, y, w, h, r, sw float32, clr color.Color) {
	strokePath(dst, roundRectPath(x, y, w, h, r), sw, clr)
}

// scaleColor умножает яркость цвета на f (для подсветки/затемнения).
func scaleColor(clr color.Color, f float64) color.RGBA {
	r, g, b, a := clr.RGBA()
	clamp := func(v float64) uint8 {
		if v > 255 {
			return 255
		}
		if v < 0 {
			return 0
		}
		return uint8(v)
	}
	return color.RGBA{
		clamp(float64(r>>8) * f),
		clamp(float64(g>>8) * f),
		clamp(float64(b>>8) * f),
		uint8(a >> 8),
	}
}

// ── Кнопка ───────────────────────────────────────────────────────────────
type Button struct {
	X, Y, W, H float32
	Label      string
	Face       text.Face
	Bg         color.Color
	Fg         color.Color
	Radius     float32 // скругление углов (0 → 12)
	Active     bool     // подсвечена (выбранный вариант)
	Disabled   bool
}

func (b *Button) contains(mx, my int) bool {
	fx, fy := float32(mx), float32(my)
	return fx >= b.X && fx <= b.X+b.W && fy >= b.Y && fy <= b.Y+b.H
}

func (b *Button) Hovered() bool {
	if b.Disabled {
		return false
	}
	mx, my := ebiten.CursorPosition()
	return b.contains(mx, my)
}

func (b *Button) Clicked() bool {
	return b.Hovered() && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
}

func (b *Button) Draw(dst *ebiten.Image) {
	r := b.Radius
	if r == 0 {
		r = 12
	}
	base := b.Bg
	if base == nil {
		base = colBtn
	}
	bg := base
	hovered := b.Hovered()
	switch {
	case b.Disabled:
		bg = color.RGBA{0x22, 0x26, 0x32, 0xff}
	case b.Active:
		bg = colBtnActive
	case hovered:
		bg = scaleColor(base, 1.22)
	}

	// Лёгкая «тень» под кнопкой для объёма.
	if !b.Disabled {
		fillRoundRect(dst, b.X, b.Y+3, b.W, b.H, r, color.RGBA{0, 0, 0, 0x55})
	}
	fillRoundRect(dst, b.X, b.Y, b.W, b.H, r, bg)
	// Верхний блик.
	fillRoundRect(dst, b.X+2, b.Y+2, b.W-4, b.H*0.42, r-2, scaleColor(bg, 1.12))
	fillRoundRect(dst, b.X+2, b.Y+b.H*0.30, b.W-4, b.H*0.68, r-2, bg)

	// Обводка: акцент при наведении/активности.
	border := scaleColor(bg, 1.4)
	if b.Active {
		border = colAccent
	} else if hovered {
		border = colAccent
	}
	strokeRoundRect(dst, b.X+1, b.Y+1, b.W-2, b.H-2, r-1, 2, border)

	fg := b.Fg
	if fg == nil {
		fg = colText
	}
	if b.Disabled {
		fg = colTextDim
	} else if b.Active {
		fg = colBg
	}
	face := b.Face
	if face == nil {
		face = faceMed
	}
	drawTextCentered(dst, b.Label, face, float64(b.X+b.W/2), float64(b.Y+b.H/2), fg)
}

// ── Утилиты ──────────────────────────────────────────────────────────────
// formatCoins добавляет разделители тысяч: 12345 → "12 345".
func formatCoins(n int) string {
	s := strconv.Itoa(n)
	neg := ""
	if n < 0 {
		neg = "-"
		s = s[1:]
	}
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ' ')
		}
		out = append(out, c)
	}
	return neg + string(out)
}

// coinStr — полная форма: «12 345 защекоинов».
func coinStr(n int) string {
	return fmt.Sprintf("%s защекоинов", formatCoins(n))
}

// coinShort — компактная форма для узких мест (плитки кейсов): «12 345 защ.».
func coinShort(n int) string {
	return fmt.Sprintf("%s защ.", formatCoins(n))
}

// textWidth возвращает ширину строки в пикселях.
func textWidth(s string, face text.Face) float64 {
	w, _ := text.Measure(s, face, 0)
	return w
}

// easeOutCubic — плавное замедление в конце анимации.
func easeOutCubic(t float64) float64 {
	t = 1 - t
	return 1 - t*t*t
}
