package mapview

import "math"

type LatLon struct {
	Lon, Lat float64
}

type Point struct{ X, Y float64 }

type BoundingBox struct{ MinX, MinY, MaxX, MaxY float64 }

func (a BoundingBox) Intersects(b BoundingBox) bool {
	return !(a.MaxX < b.MinX || a.MinX > b.MaxX || a.MaxY < b.MinY || a.MinY > b.MaxY)
}

func ComputeBounds(pts []Point) BoundingBox {
	bb := BoundingBox{math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64}
	for _, pt := range pts {
		bb.MinX = math.Min(bb.MinX, pt.X)
		bb.MaxX = math.Max(bb.MaxX, pt.X)
		bb.MinY = math.Min(bb.MinY, pt.Y)
		bb.MaxY = math.Max(bb.MaxY, pt.Y)
	}
	return bb
}

type CachedPolygon struct {
	WorldPoints []Point
	Bounds      BoundingBox
	IsHole      bool
}
