package main

import "github.com/hajimehoshi/ebiten/v2"

var chipValues = []int{10, 50, 100, 500}

// BetControl — выбор размера ставки фишками + кнопка "Всё".
type BetControl struct {
	X, Y   float32
	Amount int
	chips  []Button
	allBtn Button
}

// newBetControl раскладывает фишки + кнопку «Всё» равномерно по ширине totalW.
func newBetControl(x, y float32, start int, totalW float32) *BetControl {
	b := &BetControl{X: x, Y: y, Amount: start}
	const gap = 8
	n := len(chipValues) + 1 // фишки + «Всё»
	btnW := (totalW - float32(n-1)*gap) / float32(n)
	step := btnW + gap
	cx := x
	for _, v := range chipValues {
		b.chips = append(b.chips, Button{X: cx, Y: y, W: btnW, H: 44,
			Label: "+" + formatCoins(v), Face: faceSmall, Bg: colPanel2})
		cx += step
	}
	b.allBtn = Button{X: cx, Y: y, W: btnW, H: 44, Label: "Всё", Face: faceSmall, Bg: colPanel2}
	return b
}

// Update обрабатывает клики. balance — текущий баланс для ограничения.
// locked — ставку менять нельзя (идёт анимация).
func (b *BetControl) Update(balance int, locked bool) {
	if locked {
		return
	}
	for i := range b.chips {
		if b.chips[i].Clicked() {
			b.Amount += chipValues[i]
		}
	}
	if b.allBtn.Clicked() {
		b.Amount = balance
	}
	if b.Amount > balance {
		b.Amount = balance
	}
	if b.Amount < 0 {
		b.Amount = 0
	}
}

func (b *BetControl) reset() { b.Amount = 0 }

func (b *BetControl) Draw(dst *ebiten.Image) {
	// Строка «Ставка: N» центрируется над рядом фишек.
	label := "Ставка:  "
	amount := coinStr(b.Amount)
	wl := textWidth(label, faceSmall)
	wa := textWidth(amount, faceMed)
	center := float64(b.X+b.allBtn.X+b.allBtn.W) / 2
	startX := center - (wl+wa)/2
	drawText(dst, label, faceSmall, startX, float64(b.Y)-34, colTextDim)
	drawText(dst, amount, faceMed, startX+wl, float64(b.Y)-38, colAccent)
	for i := range b.chips {
		b.chips[i].Draw(dst)
	}
	b.allBtn.Draw(dst)
}
