
package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"shipping/internal/config"
	"shipping/internal/console"
	"shipping/internal/layer"
	"shipping/internal/mapview"
	"shipping/internal/timecontrol"
	"shipping/internal/ui"
	"shipping/internal/world"
)

type Game struct {
	Config  *config.Config
	World   *world.World
	Map     *mapview.MapView
	UI      *ui.Layer
	Time    *timecontrol.TimeControl
	Console *console.Console
}

func New(cfg *config.Config) (*Game, error) {
	mv, err := mapview.New(mapview.ConfigFromApp(cfg))
	if err != nil {
		return nil, err
	}

	w := world.NewDefault()

	mv.AddLayer(layer.NewOceanCloudLayer(
		cfg.Cloud.ZoomThreshold,
		cfg.Cloud.FadeDuration,
		cfg.Cloud.CellWorld,
		cfg.Cloud.TargetPxPerSample,
		cfg.Cloud.MaxDimension,
		cfg.Cloud.WorkerCount,
	))
	mv.AddLayer(layer.NewShippingLaneLayer(
		w,
		color.RGBA{100, 180, 220, 180},
		2.0,
	))
	mv.AddLayer(layer.NewPortLayer(
		w,
		color.RGBA{195, 160, 80, 255},
		color.RGBA{150, 125, 65, 255},
	))

	uiLayer := ui.NewLayer()

	uiLayer.AddPanel(&ui.Panel{
		ID:      "topbar",
		Visible: true,
		Anchor:  ui.AnchorTopStretch,
		OffsetX: 0,
		OffsetY: 0,
		Height:  cfg.UI.TopBarHeight,
		Padding: 12,
		Style:   ui.DefaultTopbarStyle(),
		OnDraw: func(ctx *ui.DrawContext) {
			ebitenutil.DebugPrintAt(ctx.Screen,
				cfg.Title,
				int(ctx.X), int(ctx.Y))
			ebitenutil.DebugPrintAt(ctx.Screen,
				fmt.Sprintf("%.0f FPS", ebiten.ActualFPS()),
				int(ctx.X+ctx.W-80), int(ctx.Y))
		},
	})

	uiLayer.AddPanel(&ui.Panel{
		ID:      "info_panel",
		Visible: true,
		Anchor:  ui.AnchorBottomRight,
		OffsetX: 16,
		OffsetY: 16,
		Width:   280,
		Height:  180,
		Padding: cfg.UI.PanelPadding,
		Style:   ui.DefaultPanelStyle(),
		OnDraw: func(ctx *ui.DrawContext) {
			lines := []string{
				"Selected: —",
				"",
				fmt.Sprintf("Ports:  %d", len(w.Ports)),
				fmt.Sprintf("Routes: %d active", len(w.Routes)),
			}
			y := int(ctx.Y)
			for _, line := range lines {
				ebitenutil.DebugPrintAt(ctx.Screen, line, int(ctx.X), y)
				y += 18
			}
		},
	})

	tc := timecontrol.New(cfg.Time.SpeedOptions, cfg.Time.DefaultSpeed)

	btnW := 32.0
	btnH := 28.0
	btnGap := 4.0
	btnCount := 4.0
	tcPanelW := btnCount*btnW + (btnCount-1)*btnGap + 16

	colWhite := color.RGBA{220, 220, 220, 255}
	colActive := color.RGBA{240, 200, 60, 255}

	uiLayer.AddPanel(&ui.Panel{
		ID:      "time_control",
		Visible: true,
		Anchor:  ui.AnchorTopLeft,
		OffsetX: 8,
		OffsetY: cfg.UI.TopBarHeight + 8,
		Width:   tcPanelW,
		Height:  btnH + 16,
		Padding: 8,
		Style:   ui.DefaultPanelStyle(),
		OnClick: func(ctx *ui.DrawContext, mx, my float64) {
			bx := ctx.X
			for i := 0; i < 4; i++ {
				x0 := bx + float64(i)*(btnW+btnGap)
				if mx >= x0 && mx < x0+btnW && my >= ctx.Y && my < ctx.Y+btnH {
					switch i {
					case 0:
						tc.TogglePause()
					case 1:
						tc.Paused = false
						tc.SetSpeed(1.0)
					case 2:
						tc.Paused = false
						tc.SetSpeed(5.0)
					case 3:
						tc.Paused = false
						tc.SetSpeed(25.0)
					}
				}
			}
		},
		OnDraw: func(ctx *ui.DrawContext) {
			bx := float32(ctx.X)
			by := float32(ctx.Y)
			bw := float32(btnW)
			bh := float32(btnH)
			gap := float32(btnGap)

			for i := 0; i < 4; i++ {
				x0 := bx + float32(i)*(bw+gap)
				hovered := false
				{
					mx, my := ebiten.CursorPosition()
					mxf, myf := float64(mx), float64(my)
					hovered = mxf >= float64(x0) && mxf < float64(x0+bw) &&
						myf >= float64(by) && myf < float64(by+bh)
				}

				active := false
				switch i {
				case 0:
					active = tc.Paused
				case 1:
					active = !tc.Paused && tc.Speed == 1.0
				case 2:
					active = !tc.Paused && tc.Speed == 5.0
				case 3:
					active = !tc.Paused && tc.Speed == 25.0
				}

				bgClr := color.RGBA{30, 35, 42, 200}
				if hovered {
					bgClr = color.RGBA{45, 52, 62, 220}
				}
				if active {
					bgClr = color.RGBA{50, 48, 30, 220}
				}
				vector.DrawFilledRect(ctx.Screen, x0, by, bw, bh, bgClr, false)

				iconClr := colWhite
				if active {
					iconClr = colActive
				}

				cx := x0 + bw/2
				cy := by + bh/2
				s := float32(5.0)

				switch i {
				case 0:
					vector.DrawFilledRect(ctx.Screen, cx-s+1, cy-s, s*0.4, s*2, iconClr, false)
					vector.DrawFilledRect(ctx.Screen, cx+1, cy-s, s*0.4, s*2, iconClr, false)
				case 1:
					drawTriangleRight(ctx.Screen, cx-2, cy, s, iconClr)
				case 2:
					drawTriangleRight(ctx.Screen, cx-s+1, cy, s*0.8, iconClr)
					drawTriangleRight(ctx.Screen, cx+1, cy, s*0.8, iconClr)
				case 3:
					drawTriangleRight(ctx.Screen, cx-s, cy, s*0.7, iconClr)
					drawTriangleRight(ctx.Screen, cx-2, cy, s*0.7, iconClr)
					drawTriangleRight(ctx.Screen, cx+s-4, cy, s*0.7, iconClr)
				}
			}
		},
	})

	con := console.New(cfg)

	g := &Game{Config: cfg, World: w, Map: mv, UI: uiLayer, Time: tc, Console: con}

	displayKeys := map[string]bool{
		"WINDOW_MODE": true, "WINDOW_WIDTH": true, "WINDOW_HEIGHT": true,
		"MONITOR": true, "VSYNC": true, "RESIZABLE": true, "RESOLUTION": true,
	}
	con.OnSet = func(key string) {
		if displayKeys[key] {
			g.ApplyDisplay()
		}
		mv.ApplyConfig(mapview.ConfigFromApp(cfg))
	}

	return g, nil
}

func (g *Game) Update() error {
	sw, sh := g.Map.ScreenSize()

	consoleFocused := g.Console.Update(sw, sh)

	uiFocused := false
	if !consoleFocused {
		uiFocused = g.UI.Update(sw, sh)
	}

	g.Map.SetInputSuppressed(uiFocused || consoleFocused)

	g.Map.UpdateInput()
	g.Time.Update(1.0 / 60.0)
	g.Map.UpdateLayers(g.Time)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.Map.Draw(screen)
	g.UI.Draw(screen)
	g.Console.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if outsideWidth <= 0 || outsideHeight <= 0 {
		return g.Config.Display.Resolution.Width, g.Config.Display.Resolution.Height
	}
	rh := g.Config.Display.Resolution.Height
	aspect := float64(outsideWidth) / float64(outsideHeight)
	rw := int(float64(rh)*aspect + 0.5)
	return rw, rh
}

func (g *Game) ApplyDisplay() {
	cfg := g.Config

	monitors := ebiten.AppendMonitors(nil)
	monIdx := cfg.Display.Monitor
	if monIdx < 0 || monIdx >= len(monitors) {
		monIdx = 0
	}
	if len(monitors) > 0 {
		ebiten.SetMonitor(monitors[monIdx])
	}

	ebiten.SetFullscreen(false)

	switch cfg.Display.WindowMode {
	case config.WindowModeFullscreen:
		ebiten.SetWindowDecorated(false)
		if len(monitors) > 0 {
			mw, mh := monitors[monIdx].Size()
			ebiten.SetWindowSize(mw, mh)
		}
		ebiten.SetFullscreen(true)

	case config.WindowModeBorderless:
		ebiten.SetWindowDecorated(false)
		if len(monitors) > 0 {
			mw, mh := monitors[monIdx].Size()
			ebiten.SetWindowSize(mw, mh)
		} else {
			ebiten.SetWindowSize(cfg.Display.WindowWidth, cfg.Display.WindowHeight)
		}

	default:
		ebiten.SetWindowDecorated(true)
		ebiten.SetWindowSize(cfg.Display.WindowWidth, cfg.Display.WindowHeight)
		if cfg.Display.Resizable {
			ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
		}
	}

	ebiten.SetVsyncEnabled(cfg.Display.VSync)
}

func (g *Game) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, geoM ebiten.GeoM) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM = geoM
	if g.Config.Display.FilterLinear {
		op.Filter = ebiten.FilterLinear
	}
	screen.DrawImage(offscreen, op)
}

var whitePixel *ebiten.Image

func getWhitePixel() *ebiten.Image {
	if whitePixel == nil {
		whitePixel = ebiten.NewImage(1, 1)
		whitePixel.Fill(color.White)
	}
	return whitePixel
}

func drawTriangleRight(dst *ebiten.Image, cx, cy, size float32, clr color.RGBA) {
	r, g, b, a := clr.RGBA()
	cr := float32(r) / 0xffff
	cg := float32(g) / 0xffff
	cb := float32(b) / 0xffff
	ca := float32(a) / 0xffff

	verts := []ebiten.Vertex{
		{DstX: cx - size*0.5, DstY: cy - size, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: ca},
		{DstX: cx + size, DstY: cy, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: ca},
		{DstX: cx - size*0.5, DstY: cy + size, ColorR: cr, ColorG: cg, ColorB: cb, ColorA: ca},
	}
	idx := []uint16{0, 1, 2}
	dst.DrawTriangles(verts, idx, getWhitePixel(), &ebiten.DrawTrianglesOptions{})
}
