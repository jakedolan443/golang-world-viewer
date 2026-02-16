
package mapview

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/rclancey/earcut"

	"shipping/internal/config"
	"shipping/internal/timecontrol"
)

type MapView struct {
	Cam         Camera
	Polygons    []*CachedPolygon
	Bounds      BoundingBox
	Grid        *SpatialGrid
	WhiteImg    *ebiten.Image
	ScreenW     float64
	ScreenH     float64
	Layers      []Layer
	Suppressed  bool
	ShowEquator bool
	ShowDebug   bool
	MaxZoom     float64
	Colors      mapColors
}

type mapColors struct {
	Ocean      color.RGBA
	Land       color.RGBA
	Antarctica color.RGBA
	Equator    color.RGBA
}

type Config struct {
	Shapefile    string
	RenderWidth  int
	RenderHeight int
	MaxZoom      float64
	SmoothFactor float64
	ShowEquator  bool
	ShowDebug    bool
	Colors       mapColors
}

func ConfigFromApp(cfg *config.Config) Config {
	return Config{
		Shapefile:    cfg.Map.ShapefilePath,
		RenderWidth:  cfg.Display.Resolution.Width,
		RenderHeight: cfg.Display.Resolution.Height,
		MaxZoom:      cfg.Map.MaxZoom,
		SmoothFactor: cfg.Map.SmoothFactor,
		ShowEquator:  cfg.Map.ShowEquator,
		ShowDebug:    cfg.Map.ShowDebug,
		Colors: mapColors{
			Ocean:      cfg.Map.ColorOcean,
			Land:       cfg.Map.ColorLand,
			Antarctica: cfg.Map.ColorAntarctica,
			Equator:    cfg.Map.ColorEquator,
		},
	}
}

func New(cfg Config) (*MapView, error) {
	if cfg.MaxZoom <= 0 {
		cfg.MaxZoom = 80.0
	}
	if cfg.RenderWidth <= 0 {
		cfg.RenderWidth = 2560
	}
	if cfg.RenderHeight <= 0 {
		cfg.RenderHeight = 1440
	}
	if cfg.SmoothFactor <= 0 {
		cfg.SmoothFactor = 0.15
	}

	polygons, bounds, err := LoadShapefile(cfg.Shapefile)
	if err != nil {
		return nil, fmt.Errorf("load shapefile: %w", err)
	}
	log.Printf("MapView: loaded %d polygon parts", len(polygons))

	grid := NewSpatialGrid(bounds, 10.0)
	for _, poly := range polygons {
		grid.Add(poly)
	}

	white := ebiten.NewImage(1, 1)
	white.Fill(color.White)

	sw := float64(cfg.RenderWidth)
	sh := float64(cfg.RenderHeight)
	worldW := bounds.MaxX - bounds.MinX
	worldH := bounds.MaxY - bounds.MinY
	cx := bounds.MinX + worldW/2
	cy := bounds.MinY + worldH/2
	initZoom := math.Min(sw/worldW, sh/worldH) * 0.95

	return &MapView{
		Cam:         NewCamera(cx, cy, initZoom, cfg.SmoothFactor),
		Polygons:    polygons,
		Bounds:      bounds,
		Grid:        grid,
		WhiteImg:    white,
		ScreenW:     sw,
		ScreenH:     sh,
		ShowEquator: cfg.ShowEquator,
		ShowDebug:   cfg.ShowDebug,
		MaxZoom:     cfg.MaxZoom,
		Colors:      cfg.Colors,
	}, nil
}

func (mv *MapView) AddLayer(l Layer)              { mv.Layers = append(mv.Layers, l) }
func (mv *MapView) ScreenSize() (float64, float64) { return mv.ScreenW, mv.ScreenH }
func (mv *MapView) WhiteImage() *ebiten.Image       { return mv.WhiteImg }
func (mv *MapView) CameraPos() (float64, float64)  { return mv.Cam.X, mv.Cam.Y }
func (mv *MapView) Zoom() float64                  { return mv.Cam.Zoom }
func (mv *MapView) XOffsets() []float64             { return mv.xOffsets() }

func (mv *MapView) ApplyConfig(cfg Config) {
	mv.MaxZoom = cfg.MaxZoom
	mv.ShowEquator = cfg.ShowEquator
	mv.ShowDebug = cfg.ShowDebug
	mv.Colors = cfg.Colors
	if cfg.SmoothFactor > 0 {
		mv.Cam.SmoothFactor = cfg.SmoothFactor
	}
}

func (mv *MapView) WorldBounds() (float64, float64, float64, float64) {
	return mv.Bounds.MinX, mv.Bounds.MinY, mv.Bounds.MaxX, mv.Bounds.MaxY
}

func (mv *MapView) RawWorldToScreen(lon, lat float64) (float64, float64) {
	return mv.rawWorldToScreen(lon, lat)
}

func (mv *MapView) ViewBounds() (float64, float64, float64, float64) {
	vw := mv.ScreenW / mv.Cam.Zoom
	vh := mv.ScreenH / mv.Cam.Zoom
	return mv.Cam.X - vw/2, mv.Cam.Y - vh/2, mv.Cam.X + vw/2, mv.Cam.Y + vh/2
}

func (mv *MapView) SetInputSuppressed(suppressed bool) {
	mv.Suppressed = suppressed
}

func (mv *MapView) WorldToScreen(lon, lat float64) (float64, float64) {
	ww := mv.Bounds.MaxX - mv.Bounds.MinX
	nx := lon
	for nx < mv.Bounds.MinX {
		nx += ww
	}
	for nx >= mv.Bounds.MaxX {
		nx -= ww
	}
	return mv.rawWorldToScreen(nx, lat)
}

func (mv *MapView) rawWorldToScreen(lon, lat float64) (float64, float64) {
	return (lon-mv.Cam.X)*mv.Cam.Zoom + mv.ScreenW/2,
		-(lat-mv.Cam.Y)*mv.Cam.Zoom + mv.ScreenH/2
}

func (mv *MapView) DrawCircle(screen *ebiten.Image, pos LatLon, radius float32, fill, stroke color.Color, strokeWidth float32) {
	for _, off := range mv.xOffsets() {
		sx, sy := mv.rawWorldToScreen(pos.Lon+off, pos.Lat)
		r64 := float64(radius) * 2
		if sx < -r64 || sx > mv.ScreenW+r64 || sy < -r64 || sy > mv.ScreenH+r64 {
			continue
		}
		vector.DrawFilledCircle(screen, float32(sx), float32(sy), radius, fill, false)
		if strokeWidth > 0 {
			if rgba, ok := stroke.(color.RGBA); ok {
				if rgba.A > 0 {
					vector.StrokeCircle(screen, float32(sx), float32(sy), radius, strokeWidth, stroke, false)
				}
			}
		}
	}
}

func (mv *MapView) DrawLine(screen *ebiten.Image, points []LatLon, width float32, clr color.Color) {
	if len(points) < 2 {
		return
	}
	for _, off := range mv.xOffsets() {
		for i := 0; i < len(points)-1; i++ {
			x1, y1 := mv.rawWorldToScreen(points[i].Lon+off, points[i].Lat)
			x2, y2 := mv.rawWorldToScreen(points[i+1].Lon+off, points[i+1].Lat)
			vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), width, clr, false)
		}
	}
}

func (mv *MapView) DrawPoly(screen *ebiten.Image, points []LatLon, fill, stroke color.Color, strokeWidth float32) {
	if len(points) < 3 {
		return
	}
	n := len(points)
	for _, off := range mv.xOffsets() {
		flat := make([]float64, n*2)
		for i, p := range points {
			flat[i*2], flat[i*2+1] = mv.rawWorldToScreen(p.Lon+off, p.Lat)
		}
		tri, err := earcut.Earcut(flat, nil, 2)
		if err != nil || len(tri) == 0 {
			continue
		}
		r, g, b, a := fill.RGBA()
		cr, cg, cb, ca := float32(r)/0xffff, float32(g)/0xffff, float32(b)/0xffff, float32(a)/0xffff
		verts := make([]ebiten.Vertex, n)
		for i := 0; i < n; i++ {
			verts[i] = ebiten.Vertex{
				DstX: float32(flat[i*2]), DstY: float32(flat[i*2+1]),
				ColorR: cr, ColorG: cg, ColorB: cb, ColorA: ca,
			}
		}
		idx := make([]uint16, len(tri))
		for i, ti := range tri {
			idx[i] = uint16(ti)
		}
		screen.DrawTriangles(verts, idx, mv.WhiteImg, &ebiten.DrawTrianglesOptions{})
		if strokeWidth > 0 {
			for i := 0; i < n; i++ {
				j := (i + 1) % n
				vector.StrokeLine(screen, float32(flat[i*2]), float32(flat[i*2+1]),
					float32(flat[j*2]), float32(flat[j*2+1]), strokeWidth, stroke, false)
			}
		}
	}
}

func (mv *MapView) UpdateInput() {
	ww := mv.Bounds.MaxX - mv.Bounds.MinX
	wh := mv.Bounds.MaxY - mv.Bounds.MinY

	if !mv.Suppressed {
		_, scrollY := ebiten.Wheel()

		if scrollY != 0 {
			mx, my := ebiten.CursorPosition()
			hw, hh := mv.ScreenW/2, mv.ScreenH/2
			oldWX := (float64(mx)-hw)/mv.Cam.TargetZoom + mv.Cam.TargetX
			oldWY := -(float64(my)-hh)/mv.Cam.TargetZoom + mv.Cam.TargetY

			mv.Cam.TargetZoom *= 1.0 + scrollY*0.15
			minZX := mv.ScreenW / (ww * 1.2)
			minZY := mv.ScreenH / (wh * 1.2)
			minZH := mv.ScreenH / wh
			minZ := math.Max(math.Min(minZX, minZY), minZH)
			mv.Cam.TargetZoom = math.Max(minZ, math.Min(mv.Cam.TargetZoom, mv.MaxZoom))

			newWX := (float64(mx)-hw)/mv.Cam.TargetZoom + mv.Cam.TargetX
			newWY := -(float64(my)-hh)/mv.Cam.TargetZoom + mv.Cam.TargetY
			dx := oldWX - newWX
			for dx > ww/2 {
				dx -= ww
			}
			for dx < -ww/2 {
				dx += ww
			}
			mv.Cam.TargetX += dx
			mv.Cam.TargetY += oldWY - newWY
		}

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			mx, my := ebiten.CursorPosition()
			if !mv.Cam.Dragging {
				mv.Cam.Dragging = true
				mv.Cam.LastMouseX, mv.Cam.LastMouseY = mx, my
			} else {
				dx := float64(mx-mv.Cam.LastMouseX) / mv.Cam.Zoom
				dy := float64(my-mv.Cam.LastMouseY) / mv.Cam.Zoom
				mv.Cam.TargetX -= dx
				mv.Cam.TargetY += dy
				mv.Cam.X = mv.Cam.TargetX
				mv.Cam.Y = mv.Cam.TargetY
				mv.Cam.LastMouseX, mv.Cam.LastMouseY = mx, my
			}
		} else {
			mv.Cam.Dragging = false
		}
	} else {
		mv.Cam.Dragging = false
	}

	minZH := mv.ScreenH / wh
	if mv.Cam.TargetZoom < minZH {
		mv.Cam.TargetZoom = minZH
	}
	if mv.Cam.Zoom < minZH {
		mv.Cam.Zoom = minZH
	}

	mv.Cam.Smooth()
	mv.Cam.ClampY(mv.ScreenH, mv.Bounds.MinY, mv.Bounds.MaxY)
	mv.Cam.WrapX(mv.Bounds.MinX, mv.Bounds.MaxX)
}

func (mv *MapView) UpdateLayers(tc *timecontrol.TimeControl) {
	for _, l := range mv.Layers {
		l.Update(mv, tc)
	}
}

func (mv *MapView) Draw(screen *ebiten.Image) {
	screen.Fill(mv.Colors.Ocean)
	sw := float64(screen.Bounds().Dx())
	sh := float64(screen.Bounds().Dy())
	mv.ScreenW, mv.ScreenH = sw, sh

	vw := sw / mv.Cam.Zoom
	vh := sh / mv.Cam.Zoom
	vb := BoundingBox{
		MinX: mv.Cam.X - vw/2, MaxX: mv.Cam.X + vw/2,
		MinY: mv.Cam.Y - vh/2, MaxY: mv.Cam.Y + vh/2,
	}

	ww := mv.Bounds.MaxX - mv.Bounds.MinX
	xOffs := []float64{0}
	if vb.MinX < mv.Bounds.MinX {
		xOffs = append(xOffs, -ww)
	}
	if vb.MaxX > mv.Bounds.MaxX {
		xOffs = append(xOffs, ww)
	}

	for _, l := range mv.Layers {
		if l.BelowLand() {
			l.Draw(mv, screen)
		}
	}

	rendered := 0
	for _, xOff := range xOffs {
		qb := BoundingBox{
			MinX: math.Max(vb.MinX-xOff, mv.Bounds.MinX),
			MaxX: math.Min(vb.MaxX-xOff, mv.Bounds.MaxX),
			MinY: vb.MinY, MaxY: vb.MaxY,
		}
		visible := mv.Grid.GetVisible(qb)
		for _, poly := range visible {
			if !poly.IsHole {
				mv.drawLandPolygon(screen, poly, xOff)
				rendered++
			}
		}
		for _, poly := range visible {
			if poly.IsHole {
				mv.drawLandPolygon(screen, poly, xOff)
				rendered++
			}
		}
	}

	for _, l := range mv.Layers {
		if !l.BelowLand() {
			l.Draw(mv, screen)
		}
	}

	if mv.ShowEquator {
		mv.drawEquator(screen)
	}
	if mv.ShowDebug {
		mv.drawDebug(screen, rendered, len(xOffs), vb)
	}
}

func (mv *MapView) xOffsets() []float64 {
	vw := mv.ScreenW / mv.Cam.Zoom
	ww := mv.Bounds.MaxX - mv.Bounds.MinX
	offs := []float64{0}
	if mv.Cam.X-vw/2 < mv.Bounds.MinX {
		offs = append(offs, -ww)
	}
	if mv.Cam.X+vw/2 > mv.Bounds.MaxX {
		offs = append(offs, ww)
	}
	return offs
}

func (mv *MapView) drawLandPolygon(screen *ebiten.Image, poly *CachedPolygon, xOff float64) {
	if len(poly.WorldPoints) < 3 {
		return
	}
	n := len(poly.WorldPoints)
	flat := make([]float64, n*2)
	for i, pt := range poly.WorldPoints {
		flat[i*2] = (pt.X+xOff-mv.Cam.X)*mv.Cam.Zoom + mv.ScreenW/2
		flat[i*2+1] = -(pt.Y-mv.Cam.Y)*mv.Cam.Zoom + mv.ScreenH/2
	}
	tri, err := earcut.Earcut(flat, nil, 2)
	if err != nil || len(tri) == 0 {
		return
	}

	fill := mv.Colors.Land
	if poly.Bounds.MaxY < -60.0 {
		fill = mv.Colors.Antarctica
	}
	if poly.IsHole {
		fill = mv.Colors.Ocean
	}
	r, gr, b, a := fill.RGBA()
	cr, cg, cb, ca := float32(r)/0xffff, float32(gr)/0xffff, float32(b)/0xffff, float32(a)/0xffff

	verts := make([]ebiten.Vertex, n)
	for i := 0; i < n; i++ {
		verts[i] = ebiten.Vertex{
			DstX: float32(flat[i*2]), DstY: float32(flat[i*2+1]),
			ColorR: cr, ColorG: cg, ColorB: cb, ColorA: ca,
		}
	}
	idx := make([]uint16, len(tri))
	for i, ti := range tri {
		idx[i] = uint16(ti)
	}
	screen.DrawTriangles(verts, idx, mv.WhiteImg, &ebiten.DrawTrianglesOptions{})
}

func (mv *MapView) drawEquator(screen *ebiten.Image) {
	ww := mv.Bounds.MaxX - mv.Bounds.MinX
	sx1, sy1 := mv.WorldToScreen(mv.Bounds.MinX, 0)
	sx2, sy2 := mv.WorldToScreen(mv.Bounds.MaxX, 0)
	if sy1 >= -10 && sy1 <= mv.ScreenH+10 {
		vector.StrokeLine(screen, float32(sx1), float32(sy1), float32(sx2), float32(sy2), 1.5, mv.Colors.Equator, false)
		sx3, _ := mv.WorldToScreen(mv.Bounds.MinX+ww, 0)
		sx4, _ := mv.WorldToScreen(mv.Bounds.MaxX+ww, 0)
		vector.StrokeLine(screen, float32(sx3), float32(sy1), float32(sx4), float32(sy2), 1.5, mv.Colors.Equator, false)
	}
}

func (mv *MapView) drawDebug(screen *ebiten.Image, rendered, copies int, vb BoundingBox) {
	ebitenutil.DebugPrint(screen, fmt.Sprintf(
		"FPS: %.0f | Pos: (%.1f, %.1f) Zoom: %.2f\nPolygons: %d total, %d rendered (x%d)\nView: X[%.1f..%.1f] Y[%.1f..%.1f]",
		ebiten.ActualFPS(), mv.Cam.X, mv.Cam.Y, mv.Cam.Zoom,
		len(mv.Polygons), rendered, copies, vb.MinX, vb.MaxX, vb.MinY, vb.MaxY))
}
