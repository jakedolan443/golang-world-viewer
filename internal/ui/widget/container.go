package widget

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// LayoutMode defines how widgets are arranged in a container.
type LayoutMode int

const (
	LayoutAbsolute LayoutMode = iota // Widgets use their own bounds
	LayoutHorizontal                  // Widgets arranged left-to-right
	LayoutVertical                    // Widgets arranged top-to-bottom
)

// Alignment defines how widgets align within their allocated space.
type Alignment int

const (
	AlignStart Alignment = iota // Top/Left
	AlignCenter                  // Center
	AlignEnd                     // Bottom/Right
)

// Container holds and manages child widgets.
type Container struct {
	BaseWidget
	children  []Widget
	layout    LayoutMode
	alignment Alignment
	spacing   float64
	padding   float64
}

// NewContainer creates a new container with the specified layout.
func NewContainer(x, y, w, h float64, layout LayoutMode) *Container {
	return &Container{
		BaseWidget: NewBaseWidget(x, y, w, h),
		layout:     layout,
		spacing:    4.0,
		padding:    0.0,
	}
}

// Add adds a widget to the container.
func (c *Container) Add(w Widget) {
	c.children = append(c.children, w)
	c.layoutChildren()
}

// GetChild returns the child widget at the given index, or nil if out of bounds.
func (c *Container) GetChild(index int) Widget {
	if index < 0 || index >= len(c.children) {
		return nil
	}
	return c.children[index]
}

// ChildCount returns the number of child widgets.
func (c *Container) ChildCount() int {
	return len(c.children)
}

// SetSpacing sets the spacing between widgets.
func (c *Container) SetSpacing(spacing float64) {
	c.spacing = spacing
	c.layoutChildren()
}

// SetPadding sets the padding around the container's content.
func (c *Container) SetPadding(padding float64) {
	c.padding = padding
	c.layoutChildren()
}

// SetAlignment sets how widgets align within the container.
func (c *Container) SetAlignment(align Alignment) {
	c.alignment = align
	c.layoutChildren()
}

// layoutChildren positions child widgets according to the layout mode.
func (c *Container) layoutChildren() {
	if len(c.children) == 0 {
		return
	}

	b := c.Bounds()
	contentX := b.X + c.padding
	contentY := b.Y + c.padding
	contentW := b.W - 2*c.padding
	contentH := b.H - 2*c.padding

	switch c.layout {
	case LayoutAbsolute:
		// Widgets position themselves
		for _, child := range c.children {
			cb := child.Bounds()
			child.SetBounds(Bounds{
				X: b.X + cb.X,
				Y: b.Y + cb.Y,
				W: cb.W,
				H: cb.H,
			})
		}

	case LayoutHorizontal:
		x := contentX
		for _, child := range c.children {
			cb := child.Bounds()
			y := contentY
			if c.alignment == AlignCenter {
				y = contentY + (contentH-cb.H)/2
			} else if c.alignment == AlignEnd {
				y = contentY + contentH - cb.H
			}
			child.SetBounds(Bounds{X: x, Y: y, W: cb.W, H: cb.H})
			x += cb.W + c.spacing
		}

	case LayoutVertical:
		y := contentY
		for _, child := range c.children {
			cb := child.Bounds()
			x := contentX
			if c.alignment == AlignCenter {
				x = contentX + (contentW-cb.W)/2
			} else if c.alignment == AlignEnd {
				x = contentX + contentW - cb.W
			}
			child.SetBounds(Bounds{X: x, Y: y, W: cb.W, H: cb.H})
			y += cb.H + c.spacing
		}
	}
}

// Update updates all visible children.
func (c *Container) Update() error {
	if !c.visible {
		return nil
	}
	for _, child := range c.children {
		if child.Visible() {
			if err := child.Update(); err != nil {
				return err
			}
		}
	}
	return nil
}

// Draw draws all visible children.
func (c *Container) Draw(screen *ebiten.Image) {
	if !c.visible {
		return
	}
	for _, child := range c.children {
		if child.Visible() {
			child.Draw(screen)
		}
	}
}

// SetBounds repositions the container and re-layouts children.
func (c *Container) SetBounds(bounds Bounds) {
	c.BaseWidget.SetBounds(bounds)
	c.layoutChildren()
}
