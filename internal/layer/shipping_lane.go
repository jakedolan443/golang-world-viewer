package layer

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"shipping/internal/mapview"
	"shipping/internal/timecontrol"
	"shipping/internal/world"
)

// ShippingLaneLayer renders shipping routes from the world model.
type ShippingLaneLayer struct {
	world     *world.World
	lineColor color.RGBA
	lineWidth float32
}

func NewShippingLaneLayer(w *world.World, lineColor color.RGBA, lineWidth float32) *ShippingLaneLayer {
	return &ShippingLaneLayer{world: w, lineColor: lineColor, lineWidth: lineWidth}
}

func (s *ShippingLaneLayer) BelowLand() bool                                        { return false }
func (s *ShippingLaneLayer) Update(_ *mapview.MapView, _ *timecontrol.TimeControl)   {}

func (s *ShippingLaneLayer) Draw(mv *mapview.MapView, screen *ebiten.Image) {
	for _, route := range s.world.Routes {
		if len(route.Waypoints) < 2 {
			continue
		}
		pts := make([]mapview.LatLon, len(route.Waypoints))
		for i, wp := range route.Waypoints {
			pts[i] = mapview.LatLon{Lon: wp[0], Lat: wp[1]}
		}
		mv.DrawLine(screen, pts, s.lineWidth, s.lineColor)
	}
}
