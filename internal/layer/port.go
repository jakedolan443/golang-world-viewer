package layer

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"shipping/internal/mapview"
	"shipping/internal/timecontrol"
	"shipping/internal/world"
)

// PortLayer renders world ports as circles on the map.
type PortLayer struct {
	world  *world.World
	fill   color.RGBA
	stroke color.RGBA
}

func NewPortLayer(w *world.World, fill, stroke color.RGBA) *PortLayer {
	return &PortLayer{world: w, fill: fill, stroke: stroke}
}

func (p *PortLayer) BelowLand() bool                                        { return false }
func (p *PortLayer) Update(_ *mapview.MapView, _ *timecontrol.TimeControl)   {}

func (p *PortLayer) Draw(mv *mapview.MapView, screen *ebiten.Image) {
	zoom := mv.Zoom()
	radius := float32(math.Max(2.5, math.Min(8, 4*zoom)))

	for _, port := range p.world.Ports {
		pos := mapview.LatLon{Lon: port.Lon, Lat: port.Lat}
		mv.DrawCircle(screen, pos, radius, p.fill, p.stroke, 1.0)
	}
}
