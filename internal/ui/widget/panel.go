package widget

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Panel is a colored background rectangle, often used as a container background.
type Panel struct {
	BaseWidget
	background  color.RGBA
	border      color.RGBA
	borderWidth float32
	radius      float32
}

// NewPanel creates a new panel with the given colors.
func NewPanel(x, y, w, h float64, background, border color.RGBA) *Panel {
	return &Panel{
		BaseWidget:  NewBaseWidget(x, y, w, h),
		background:  background,
		border:      border,
		borderWidth: 1.0,
		radius:      6.0,
	}
}

// SetBackground sets the panel background color.
func (p *Panel) SetBackground(c color.RGBA) {
	p.background = c
}

// SetBorder sets the panel border color and width.
func (p *Panel) SetBorder(c color.RGBA, width float32) {
	p.border = c
	p.borderWidth = width
}

// SetRadius sets the corner radius.
func (p *Panel) SetRadius(radius float32) {
	p.radius = radius
}

// Draw renders the panel.
func (p *Panel) Draw(screen *ebiten.Image) {
	if !p.visible {
		return
	}

	bounds := p.Bounds()
	x, y := float32(bounds.X), float32(bounds.Y)
	w, h := float32(bounds.W), float32(bounds.H)

	// Draw background
	if p.background.A > 0 {
		if p.radius > 0 {
			drawRoundedRect(screen, x, y, w, h, p.radius, p.background)
		} else {
			vector.DrawFilledRect(screen, x, y, w, h, p.background, false)
		}
	}

	// Draw border
	if p.borderWidth > 0 && p.border.A > 0 {
		if p.radius > 0 {
			strokeRoundedRect(screen, x, y, w, h, p.radius, p.borderWidth, p.border)
		} else {
			vector.StrokeRect(screen, x, y, w, h, p.borderWidth, p.border, false)
		}
	}
}
