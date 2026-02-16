
package game

import (
	"fmt"
	"image/color"

	"github.com/ebitenui/ebitenui"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"shipping/internal/config"
	"shipping/internal/console"
	"shipping/internal/layer"
	"shipping/internal/mapview"
	"shipping/internal/timecontrol"
	"shipping/internal/ui"
	"shipping/internal/world"
)

type Game struct {
	Config   *config.Config
	World    *world.World
	Map      *mapview.MapView
	UI       *ui.Layer
	EbitenUI *ebitenui.UI
	Time     *timecontrol.TimeControl
	Console  *console.Console

	fpsLabel *widget.Text
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

	eui, fpsLabel := buildTopbar(cfg, tc)

	con := console.New(cfg)

	g := &Game{
		Config:   cfg,
		World:    w,
		Map:      mv,
		UI:       uiLayer,
		EbitenUI: eui,
		Time:     tc,
		Console:  con,
		fpsLabel: fpsLabel,
	}

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

func buildTopbar(cfg *config.Config, tc *timecontrol.TimeControl) (*ebitenui.UI, *widget.Text) {
	face := ui.Face()
	titleFace := ui.FaceSize(16)
	white := color.NRGBA{220, 220, 220, 255}
	topbarBg := color.NRGBA{16, 19, 25, 230}

	root := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	topbar := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(euiimage.NewNineSliceColor(topbarBg)),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				StretchHorizontal: true,
			}),
			widget.WidgetOpts.MinSize(0, int(cfg.UI.TopBarHeight)),
		),
	)

	// Left: title
	titleLabel := widget.NewText(
		widget.TextOpts.Text(cfg.Title, &titleFace, white),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionStart,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
				Padding:            &widget.Insets{Left: 16},
			}),
		),
	)

	// Center: sample colored buttons
	centerSection := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(8),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)
	centerSection.AddChild(ui.TextButton("Trade", ui.ButtonColors(color.NRGBA{40, 80, 160, 255}), white, face, nil))
	centerSection.AddChild(ui.TextButton("Fleet", ui.ButtonColors(color.NRGBA{40, 140, 60, 255}), white, face, nil))
	centerSection.AddChild(ui.TextButton("Events", ui.ButtonColors(color.NRGBA{160, 50, 50, 255}), white, face, nil))

	// Right: time controls + FPS
	rightSection := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(4),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionEnd,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
				Padding:            &widget.Insets{Right: 16},
			}),
		),
	)

	tcBase := color.NRGBA{30, 35, 42, 220}
	rightSection.AddChild(ui.TextButton("||", ui.ButtonColors(tcBase), white, face, func() {
		tc.TogglePause()
	}))
	rightSection.AddChild(ui.TextButton("1x", ui.ButtonColors(tcBase), white, face, func() {
		tc.Paused = false
		tc.SetSpeed(1.0)
	}))
	rightSection.AddChild(ui.TextButton("5x", ui.ButtonColors(tcBase), white, face, func() {
		tc.Paused = false
		tc.SetSpeed(5.0)
	}))
	rightSection.AddChild(ui.TextButton("25x", ui.ButtonColors(tcBase), white, face, func() {
		tc.Paused = false
		tc.SetSpeed(25.0)
	}))

	fpsLabel := widget.NewText(
		widget.TextOpts.Text("0 FPS", &face, color.NRGBA{160, 160, 160, 255}),
	)
	rightSection.AddChild(fpsLabel)

	topbar.AddChild(titleLabel)
	topbar.AddChild(centerSection)
	topbar.AddChild(rightSection)
	root.AddChild(topbar)

	return ui.NewUI(root), fpsLabel
}

func (g *Game) Update() error {
	sw, sh := g.Map.ScreenSize()

	consoleFocused := g.Console.Update(sw, sh)

	g.EbitenUI.Update()

	uiFocused := false
	if !consoleFocused {
		uiFocused = g.UI.Update(sw, sh)
	}

	// Suppress map input when cursor is over topbar or other UI
	_, my := ebiten.CursorPosition()
	topbarHovered := float64(my) < g.Config.UI.TopBarHeight

	g.Map.SetInputSuppressed(uiFocused || consoleFocused || topbarHovered)

	g.Map.UpdateInput()
	g.Time.Update(1.0 / 60.0)
	g.Map.UpdateLayers(g.Time)

	g.fpsLabel.Label = fmt.Sprintf("%.0f FPS", ebiten.ActualFPS())

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.Map.Draw(screen)
	g.UI.Draw(screen)
	g.EbitenUI.Draw(screen)
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
