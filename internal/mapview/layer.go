package mapview

import (
	"github.com/hajimehoshi/ebiten/v2"

	"shipping/internal/timecontrol"
)

type Layer interface {

	Update(mv *MapView, tc *timecontrol.TimeControl)

	Draw(mv *MapView, screen *ebiten.Image)

	BelowLand() bool
}
