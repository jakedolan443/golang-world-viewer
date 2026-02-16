package widget

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Bounds represents a rectangular area.
type Bounds struct {
	X, Y, W, H float64
}

// Contains checks if a point is within the bounds.
func (b Bounds) Contains(x, y float64) bool {
	return x >= b.X && x < b.X+b.W && y >= b.Y && y < b.Y+b.H
}

// Widget is the base interface for all UI elements.
type Widget interface {
	// Update handles input and state changes.
	Update() error

	// Draw renders the widget to the screen.
	Draw(screen *ebiten.Image)

	// Bounds returns the widget's current position and size.
	Bounds() Bounds

	// SetBounds sets the widget's position and size.
	SetBounds(b Bounds)

	// Visible returns whether the widget should be drawn and handle input.
	Visible() bool

	// SetVisible sets the visibility state.
	SetVisible(visible bool)
}

// BaseWidget provides common functionality for all widgets.
type BaseWidget struct {
	bounds  Bounds
	visible bool
}

func NewBaseWidget(x, y, w, h float64) BaseWidget {
	return BaseWidget{
		bounds:  Bounds{X: x, Y: y, W: w, H: h},
		visible: true,
	}
}

func (b *BaseWidget) Bounds() Bounds              { return b.bounds }
func (b *BaseWidget) SetBounds(bounds Bounds)     { b.bounds = bounds }
func (b *BaseWidget) Visible() bool               { return b.visible }
func (b *BaseWidget) SetVisible(visible bool)     { b.visible = visible }
func (b *BaseWidget) Update() error                { return nil }
func (b *BaseWidget) Draw(screen *ebiten.Image)  {}
