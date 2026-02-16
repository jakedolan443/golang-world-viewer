package widget

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ButtonState represents the current state of a button.
type ButtonState int

const (
	ButtonNormal ButtonState = iota
	ButtonHover
	ButtonPressed
)

// Button is a clickable UI element.
type Button struct {
	BaseWidget
	text       string
	onClick    func()
	state      ButtonState

	// Styling
	colorNormal  color.RGBA
	colorHover   color.RGBA
	colorPressed color.RGBA
	colorText    color.RGBA
	colorBorder  color.RGBA
	borderWidth  float32
	radius       float32
}

// NewButton creates a new button with the given text and click handler.
func NewButton(x, y, w, h float64, text string, onClick func()) *Button {
	return &Button{
		BaseWidget:   NewBaseWidget(x, y, w, h),
		text:         text,
		onClick:      onClick,
		state:        ButtonNormal,
		colorNormal:  color.RGBA{45, 52, 62, 220},
		colorHover:   color.RGBA{60, 70, 82, 230},
		colorPressed: color.RGBA{35, 42, 50, 240},
		colorText:    color.RGBA{220, 220, 220, 255},
		colorBorder:  color.RGBA{80, 90, 105, 200},
		borderWidth:  1.0,
		radius:       4.0,
	}
}

// SetColors sets all button colors at once.
func (b *Button) SetColors(normal, hover, pressed, text, border color.RGBA) {
	b.colorNormal = normal
	b.colorHover = hover
	b.colorPressed = pressed
	b.colorText = text
	b.colorBorder = border
}

// SetRadius sets the corner radius.
func (b *Button) SetRadius(radius float32) {
	b.radius = radius
}

// SetText changes the button text.
func (b *Button) SetText(text string) {
	b.text = text
}

// Update handles mouse interaction.
func (b *Button) Update() error {
	if !b.visible {
		return nil
	}

	mx, my := ebiten.CursorPosition()
	mxf, myf := float64(mx), float64(my)
	hovered := b.bounds.Contains(mxf, myf)

	if hovered {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			b.state = ButtonPressed
		} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && b.state == ButtonPressed {
			// Stay pressed
		} else {
			if b.state == ButtonPressed && b.onClick != nil {
				b.onClick()
			}
			b.state = ButtonHover
		}
	} else {
		b.state = ButtonNormal
	}

	return nil
}

// Draw renders the button.
func (b *Button) Draw(screen *ebiten.Image) {
	if !b.visible {
		return
	}

	bounds := b.Bounds()
	x, y := float32(bounds.X), float32(bounds.Y)
	w, h := float32(bounds.W), float32(bounds.H)

	// Select background color based on state
	var bgColor color.RGBA
	switch b.state {
	case ButtonPressed:
		bgColor = b.colorPressed
	case ButtonHover:
		bgColor = b.colorHover
	default:
		bgColor = b.colorNormal
	}

	// Draw background
	if b.radius > 0 {
		drawRoundedRect(screen, x, y, w, h, b.radius, bgColor)
	} else {
		vector.DrawFilledRect(screen, x, y, w, h, bgColor, false)
	}

	// Draw border
	if b.borderWidth > 0 && b.colorBorder.A > 0 {
		if b.radius > 0 {
			strokeRoundedRect(screen, x, y, w, h, b.radius, b.borderWidth, b.colorBorder)
		} else {
			vector.StrokeRect(screen, x, y, w, h, b.borderWidth, b.colorBorder, false)
		}
	}

	// Draw text (centered)
	if b.text != "" {
		// Approximate text width (6 pixels per char)
		textW := len(b.text) * 6
		textH := 16
		textX := int(bounds.X + (bounds.W-float64(textW))/2)
		textY := int(bounds.Y + (bounds.H-float64(textH))/2 + 2)
		ebitenutil.DebugPrintAt(screen, b.text, textX, textY)
	}
}

// Helper functions for rounded rectangles
func drawRoundedRect(dst *ebiten.Image, x, y, w, h, r float32, clr color.RGBA) {
	if r > w/2 {
		r = w / 2
	}
	if r > h/2 {
		r = h / 2
	}
	// Main body
	vector.DrawFilledRect(dst, x+r, y, w-2*r, h, clr, false)
	vector.DrawFilledRect(dst, x, y+r, w, h-2*r, clr, false)
	// Corners
	vector.DrawFilledCircle(dst, x+r, y+r, r, clr, false)
	vector.DrawFilledCircle(dst, x+w-r, y+r, r, clr, false)
	vector.DrawFilledCircle(dst, x+r, y+h-r, r, clr, false)
	vector.DrawFilledCircle(dst, x+w-r, y+h-r, r, clr, false)
}

func strokeRoundedRect(dst *ebiten.Image, x, y, w, h, r, bw float32, clr color.RGBA) {
	if r > w/2 {
		r = w / 2
	}
	if r > h/2 {
		r = h / 2
	}
	// Edges
	vector.StrokeLine(dst, x+r, y, x+w-r, y, bw, clr, false)
	vector.StrokeLine(dst, x+r, y+h, x+w-r, y+h, bw, clr, false)
	vector.StrokeLine(dst, x, y+r, x, y+h-r, bw, clr, false)
	vector.StrokeLine(dst, x+w, y+r, x+w, y+h-r, bw, clr, false)

	// Corner arcs (simplified - using line segments)
	const segs = 6
	for i := 0; i < segs; i++ {
		// Top-left
		a0 := float64(i) * 3.14159 / (2 * segs)
		a1 := float64(i+1) * 3.14159 / (2 * segs)
		vector.StrokeLine(dst,
			x+r-r*float32(cos(a0+3.14159)), y+r-r*float32(sin(a0+3.14159)),
			x+r-r*float32(cos(a1+3.14159)), y+r-r*float32(sin(a1+3.14159)),
			bw, clr, false)
		// Top-right
		vector.StrokeLine(dst,
			x+w-r+r*float32(cos(a0+3.14159/2)), y+r-r*float32(sin(a0+3.14159/2)),
			x+w-r+r*float32(cos(a1+3.14159/2)), y+r-r*float32(sin(a1+3.14159/2)),
			bw, clr, false)
		// Bottom-left
		vector.StrokeLine(dst,
			x+r-r*float32(cos(a0+3.14159/2)), y+h-r+r*float32(sin(a0+3.14159/2)),
			x+r-r*float32(cos(a1+3.14159/2)), y+h-r+r*float32(sin(a1+3.14159/2)),
			bw, clr, false)
		// Bottom-right
		vector.StrokeLine(dst,
			x+w-r+r*float32(cos(a0)), y+h-r+r*float32(sin(a0)),
			x+w-r+r*float32(cos(a1)), y+h-r+r*float32(sin(a1)),
			bw, clr, false)
	}
}

func cos(a float64) float64 {
	// Simple approximation for small angles
	return 1 - a*a/2 + a*a*a*a/24
}

func sin(a float64) float64 {
	return a - a*a*a/6 + a*a*a*a*a/120
}
