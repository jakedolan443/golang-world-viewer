package mapview

import "math"

const maxPointsForDirectRender = 500

func SubdivideByGrid(pts []Point, cellSize float64) [][]Point {
	bb := ComputeBounds(pts)
	sx := math.Floor(bb.MinX/cellSize) * cellSize
	ex := math.Ceil(bb.MaxX/cellSize) * cellSize
	sy := math.Floor(bb.MinY/cellSize) * cellSize
	ey := math.Ceil(bb.MaxY/cellSize) * cellSize

	var result [][]Point
	for y := sy; y < ey; y += cellSize {
		for x := sx; x < ex; x += cellSize {
			if c := ClipToRect(pts, x, x+cellSize, y, y+cellSize); len(c) >= 3 {
				result = append(result, c)
			}
		}
	}
	return result
}

func ClipToRect(pts []Point, minX, maxX, minY, maxY float64) []Point {
	c := clipEdge(pts,
		func(p Point) bool { return p.Y >= minY },
		func(a, b Point) Point { t := (minY - a.Y) / (b.Y - a.Y); return Point{a.X + t*(b.X-a.X), minY} })
	c = clipEdge(c,
		func(p Point) bool { return p.Y <= maxY },
		func(a, b Point) Point { t := (maxY - a.Y) / (b.Y - a.Y); return Point{a.X + t*(b.X-a.X), maxY} })
	c = clipEdge(c,
		func(p Point) bool { return p.X >= minX },
		func(a, b Point) Point { t := (minX - a.X) / (b.X - a.X); return Point{minX, a.Y + t*(b.Y-a.Y)} })
	c = clipEdge(c,
		func(p Point) bool { return p.X <= maxX },
		func(a, b Point) Point { t := (maxX - a.X) / (b.X - a.X); return Point{maxX, a.Y + t*(b.Y-a.Y)} })
	return c
}

func clipEdge(pts []Point, inside func(Point) bool, intersect func(Point, Point) Point) []Point {
	if len(pts) == 0 {
		return nil
	}
	var out []Point
	n := len(pts)
	for i := 0; i < n; i++ {
		cur, nxt := pts[i], pts[(i+1)%n]
		cIn, nIn := inside(cur), inside(nxt)
		if cIn {
			out = append(out, cur)
			if !nIn {
				out = append(out, intersect(cur, nxt))
			}
		} else if nIn {
			out = append(out, intersect(cur, nxt))
		}
	}
	return out
}

func IsHoleRing(pts []Point) bool {
	area := 0.0
	n := len(pts)
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		area += pts[i].X*pts[j].Y - pts[j].X*pts[i].Y
	}
	return area > 0
}
