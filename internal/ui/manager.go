package ui

import (
	"github.com/hajimehoshi/ebiten/v2"

	"shipping/internal/ui/widget"
)

// Manager manages the UI widget hierarchy.
type Manager struct {
	root   *widget.Container
	theme  *Theme
	screenW float64
	screenH float64
}

// NewManager creates a new UI manager.
func NewManager(theme *Theme) *Manager {
	return &Manager{
		root:  widget.NewContainer(0, 0, 800, 600, widget.LayoutAbsolute),
		theme: theme,
	}
}

// Root returns the root container for adding widgets.
func (m *Manager) Root() *widget.Container {
	return m.root
}

// Theme returns the current theme.
func (m *Manager) Theme() *Theme {
	return m.theme
}

// Update updates all widgets and returns true if any widget captured input.
func (m *Manager) Update(screenW, screenH float64) bool {
	m.screenW = screenW
	m.screenH = screenH
	m.root.SetBounds(widget.Bounds{X: 0, Y: 0, W: screenW, H: screenH})

	if err := m.root.Update(); err != nil {
		// Log error if needed
		_ = err
	}

	// For now, assume UI captures input if mouse is over any widget
	// This could be improved with a proper focus/capture system
	return false
}

// Draw renders all widgets.
func (m *Manager) Draw(screen *ebiten.Image) {
	m.root.Draw(screen)
}
