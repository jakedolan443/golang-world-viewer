package mapview

import (
	"fmt"
	"math"

	"github.com/jonas-p/go-shp"
)

func LoadShapefile(path string) ([]*CachedPolygon, BoundingBox, error) {
	shape, err := shp.Open(path)
	if err != nil {
		return nil, BoundingBox{}, fmt.Errorf("open shapefile: %w", err)
	}
	defer shape.Close()

	var polygons []*CachedPolygon
	bounds := BoundingBox{math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64}

	for shape.Next() {
		_, p := shape.Shape()
		s, ok := p.(*shp.Polygon)
		if !ok {
			continue
		}
		for pi := 0; pi < len(s.Parts); pi++ {
			start := s.Parts[pi]
			end := s.NumPoints
			if pi < len(s.Parts)-1 {
				end = s.Parts[pi+1]
			}

			var pts []Point
			pb := BoundingBox{math.MaxFloat64, math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64}
			for i := start; i < end; i++ {
				pt := Point{s.Points[i].X, s.Points[i].Y}
				pts = append(pts, pt)
				pb.MinX = math.Min(pb.MinX, pt.X)
				pb.MaxX = math.Max(pb.MaxX, pt.X)
				pb.MinY = math.Min(pb.MinY, pt.Y)
				pb.MaxY = math.Max(pb.MaxY, pt.Y)
				bounds.MinX = math.Min(bounds.MinX, pt.X)
				bounds.MaxX = math.Max(bounds.MaxX, pt.X)
				bounds.MinY = math.Min(bounds.MinY, pt.Y)
				bounds.MaxY = math.Max(bounds.MaxY, pt.Y)
			}
			if len(pts) < 3 {
				continue
			}

			hole := IsHoleRing(pts)
			if len(pts) > maxPointsForDirectRender {
				for _, sub := range SubdivideByGrid(pts, 10.0) {
					if len(sub) >= 3 {
						polygons = append(polygons, &CachedPolygon{
							WorldPoints: sub,
							Bounds:      ComputeBounds(sub),
							IsHole:      hole,
						})
					}
				}
			} else {
				polygons = append(polygons, &CachedPolygon{
					WorldPoints: pts,
					Bounds:      pb,
					IsHole:      hole,
				})
			}
		}
	}
	return polygons, bounds, nil
}
