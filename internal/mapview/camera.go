package mapview

import "math"

type Camera struct {
	X, Y           float64
	Zoom           float64
	TargetX        float64
	TargetY        float64
	TargetZoom     float64
	Dragging       bool
	LastMouseX     int
	LastMouseY     int
	SmoothFactor   float64
}

func NewCamera(cx, cy, zoom, smoothFactor float64) Camera {
	return Camera{
		X: cx, Y: cy, TargetX: cx, TargetY: cy,
		Zoom: zoom, TargetZoom: zoom,
		SmoothFactor: smoothFactor,
	}
}

func (c *Camera) Smooth() {
	sf := c.SmoothFactor

	if math.Abs(c.Zoom-c.TargetZoom) > 0.0001 {
		c.Zoom += (c.TargetZoom - c.Zoom) * sf
	} else {
		c.Zoom = c.TargetZoom
	}

	if math.Abs(c.X-c.TargetX) > 0.0001 || math.Abs(c.Y-c.TargetY) > 0.0001 {
		c.X += (c.TargetX - c.X) * sf
		c.Y += (c.TargetY - c.Y) * sf
	} else {
		c.X = c.TargetX
		c.Y = c.TargetY
	}
}

func (c *Camera) ClampY(screenH, worldMinY, worldMaxY float64) {
	vh := screenH / c.Zoom
	minY := worldMinY + vh/2
	maxY := worldMaxY - vh/2
	if minY > maxY {
		mid := (worldMinY + worldMaxY) / 2
		c.TargetY, c.Y = mid, mid
		return
	}
	c.TargetY = math.Max(minY, math.Min(maxY, c.TargetY))
	c.Y = math.Max(minY, math.Min(maxY, c.Y))
}

func (c *Camera) WrapX(minX, maxX float64) {
	ww := maxX - minX
	for c.TargetX < minX {
		c.TargetX += ww
	}
	for c.TargetX >= maxX {
		c.TargetX -= ww
	}
	for c.X < minX {
		c.X += ww
	}
	for c.X >= maxX {
		c.X -= ww
	}
}
