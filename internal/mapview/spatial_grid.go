package mapview

import "math"

type SpatialGrid struct {
	cellSize float64
	cells    map[int]map[int][]*CachedPolygon
	bounds   BoundingBox

	// Reusable buffers to avoid per-frame allocations.
	seen   map[*CachedPolygon]bool
	result []*CachedPolygon
}

func NewSpatialGrid(bounds BoundingBox, cellSize float64) *SpatialGrid {
	return &SpatialGrid{
		cellSize: cellSize,
		cells:    make(map[int]map[int][]*CachedPolygon),
		bounds:   bounds,
		seen:     make(map[*CachedPolygon]bool),
	}
}

func (sg *SpatialGrid) Add(poly *CachedPolygon) {
	minCX := int(math.Floor(poly.Bounds.MinX / sg.cellSize))
	maxCX := int(math.Ceil(poly.Bounds.MaxX / sg.cellSize))
	minCY := int(math.Floor(poly.Bounds.MinY / sg.cellSize))
	maxCY := int(math.Ceil(poly.Bounds.MaxY / sg.cellSize))
	for x := minCX; x <= maxCX; x++ {
		if sg.cells[x] == nil {
			sg.cells[x] = make(map[int][]*CachedPolygon)
		}
		for y := minCY; y <= maxCY; y++ {
			sg.cells[x][y] = append(sg.cells[x][y], poly)
		}
	}
}

func (sg *SpatialGrid) GetVisible(vb BoundingBox) []*CachedPolygon {
	minCX := int(math.Floor(vb.MinX / sg.cellSize))
	maxCX := int(math.Ceil(vb.MaxX / sg.cellSize))
	minCY := int(math.Floor(vb.MinY / sg.cellSize))
	maxCY := int(math.Ceil(vb.MaxY / sg.cellSize))

	clear(sg.seen)
	sg.result = sg.result[:0]

	for x := minCX; x <= maxCX; x++ {
		col := sg.cells[x]
		if col == nil {
			continue
		}
		for y := minCY; y <= maxCY; y++ {
			for _, p := range col[y] {
				if !sg.seen[p] && p.Bounds.Intersects(vb) {
					sg.seen[p] = true
					sg.result = append(sg.result, p)
				}
			}
		}
	}
	return sg.result
}
