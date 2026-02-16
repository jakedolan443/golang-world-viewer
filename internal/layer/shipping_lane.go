package layer

import (
	"github.com/hajimehoshi/ebiten/v2"

	"shipping/internal/mapview"
	"shipping/internal/timecontrol"
)

type ShippingLaneLayer struct{}

func NewShippingLaneLayer() *ShippingLaneLayer { return &ShippingLaneLayer{} }

func (s *ShippingLaneLayer) BelowLand() bool                                         { return false }
func (s *ShippingLaneLayer) Update(_ *mapview.MapView, _ *timecontrol.TimeControl)    {}
func (s *ShippingLaneLayer) Draw(_ *mapview.MapView, _ *ebiten.Image)                 {}
