package widget

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// TextAlign defines how text is aligned within its bounds.
type TextAlign int

const (
	AlignLeft TextAlign = iota
	AlignCenterText
	AlignRight
)

// Label displays static text.
type Label struct {
	BaseWidget
	text  string
	color color.RGBA
	align TextAlign
}

// NewLabel creates a new label with the given text.
func NewLabel(x, y, w, h float64, text string) *Label {
	return &Label{
		BaseWidget: NewBaseWidget(x, y, w, h),
		text:       text,
		color:      color.RGBA{220, 220, 220, 255},
		align:      AlignLeft,
	}
}

// SetText changes the label text.
func (l *Label) SetText(text string) {
	l.text = text
}

// SetColor sets the text color.
func (l *Label) SetColor(c color.RGBA) {
	l.color = c
}

// SetAlign sets the text alignment.
func (l *Label) SetAlign(align TextAlign) {
	l.align = align
}

// Draw renders the label.
func (l *Label) Draw(screen *ebiten.Image) {
	if !l.visible || l.text == "" {
		return
	}

	bounds := l.Bounds()
	textW := len(l.text) * 6 // Approximate width
	textH := 16

	var textX int
	switch l.align {
	case AlignLeft:
		textX = int(bounds.X)
	case AlignCenterText:
		textX = int(bounds.X + (bounds.W-float64(textW))/2)
	case AlignRight:
		textX = int(bounds.X + bounds.W - float64(textW))
	}

	textY := int(bounds.Y + (bounds.H-float64(textH))/2 + 2)
	ebitenutil.DebugPrintAt(screen, l.text, textX, textY)
}
