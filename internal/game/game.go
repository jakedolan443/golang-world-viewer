
package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"shipping/internal/config"
	"shipping/internal/console"
	"shipping/internal/layer"
	"shipping/internal/mapview"
	"shipping/internal/timecontrol"
	"shipping/internal/ui"
	"shipping/internal/ui/widget"
	"shipping/internal/world"
)

type Game struct {
	Config     *config.Config
	World      *world.World
	Map        *mapview.MapView
	UIManager  *ui.Manager
	OldPanels  *ui.Layer // Keep for info panel
	Time       *timecontrol.TimeControl
	Console    *console.Console

	// FPS label for dynamic updates
	fpsLabel *widget.Label
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

	// Create new UI framework manager
	theme := ui.DefaultTheme()
	uiMgr := ui.NewManager(theme)

	tc := timecontrol.New(cfg.Time.SpeedOptions, cfg.Time.DefaultSpeed)

	// Build the topbar using the new widget system
	buildTopBar(uiMgr, cfg, tc)

	// Keep old system for info panel (bottom-right)
	oldPanels := ui.NewLayer()
	oldPanels.AddPanel(&ui.Panel{
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

	con := console.New(cfg)

	// Find FPS label for updates
	var fpsLabel *widget.Label
	// We'll set this in buildTopBar by returning it

	g := &Game{
		Config:    cfg,
		World:     w,
		Map:       mv,
		UIManager: uiMgr,
		OldPanels: oldPanels,
		Time:      tc,
		Console:   con,
		fpsLabel:  fpsLabel, // Will be set in buildTopBar
	}

	// Actually rebuild with access to game struct
	g.buildTopBar()

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

func (g *Game) buildTopBar() {
	root := g.UIManager.Root()
	theme := g.UIManager.Theme()
	cfg := g.Config
	tc := g.Time

	// Create topbar background panel
	topBarHeight := cfg.UI.TopBarHeight
	topBar := widget.NewPanel(0, 0, 100, topBarHeight, theme.PanelBackground, theme.PanelBorder)
	topBar.SetRadius(0) // No rounded corners for topbar

	// Create a container for topbar content
	topBarContainer := widget.NewContainer(0, 0, 100, topBarHeight, widget.LayoutAbsolute)

	// Title label (left)
	titleLabel := widget.NewLabel(theme.PaddingLarge, 0, 200, topBarHeight, cfg.Title)
	topBarContainer.Add(titleLabel)

	// FPS label (left, after title)
	g.fpsLabel = widget.NewLabel(theme.PaddingLarge+200, 0, 100, topBarHeight, "60 FPS")
	topBarContainer.Add(g.fpsLabel)

	// Sample colored buttons in the middle
	middleButtonsContainer := widget.NewContainer(0, 0, 300, topBarHeight, widget.LayoutHorizontal)
	middleButtonsContainer.SetSpacing(theme.Spacing)
	middleButtonsContainer.SetPadding(theme.PaddingMedium)
	middleButtonsContainer.SetAlignment(widget.AlignCenter)

	// Button 1: Blue
	btn1Normal, btn1Hover, btn1Pressed := ui.ColoredButtonTheme(color.RGBA{60, 120, 200, 220})
	btn1 := widget.NewButton(0, 0, 70, 32, "Blue", func() {
		fmt.Println("Blue button clicked!")
	})
	btn1.SetColors(btn1Normal, btn1Hover, btn1Pressed, theme.ButtonText, theme.ButtonBorder)
	middleButtonsContainer.Add(btn1)

	// Button 2: Green
	btn2Normal, btn2Hover, btn2Pressed := ui.ColoredButtonTheme(color.RGBA{60, 180, 100, 220})
	btn2 := widget.NewButton(0, 0, 70, 32, "Green", func() {
		fmt.Println("Green button clicked!")
	})
	btn2.SetColors(btn2Normal, btn2Hover, btn2Pressed, theme.ButtonText, theme.ButtonBorder)
	middleButtonsContainer.Add(btn2)

	// Button 3: Red
	btn3Normal, btn3Hover, btn3Pressed := ui.ColoredButtonTheme(color.RGBA{200, 80, 80, 220})
	btn3 := widget.NewButton(0, 0, 70, 32, "Red", func() {
		fmt.Println("Red button clicked!")
	})
	btn3.SetColors(btn3Normal, btn3Hover, btn3Pressed, theme.ButtonText, theme.ButtonBorder)
	middleButtonsContainer.Add(btn3)

	topBarContainer.Add(middleButtonsContainer)

	// Time control buttons (right side)
	timeControlContainer := widget.NewContainer(0, 0, 160, topBarHeight, widget.LayoutHorizontal)
	timeControlContainer.SetSpacing(theme.Spacing)
	timeControlContainer.SetPadding(theme.PaddingMedium)
	timeControlContainer.SetAlignment(widget.AlignCenter)

	// Pause button
	pauseBtn := widget.NewButton(0, 0, 36, 32, "||", func() {
		tc.TogglePause()
	})
	timeControlContainer.Add(pauseBtn)

	// 1x speed
	speed1Btn := widget.NewButton(0, 0, 36, 32, "1x", func() {
		tc.Paused = false
		tc.SetSpeed(1.0)
	})
	timeControlContainer.Add(speed1Btn)

	// 5x speed
	speed5Btn := widget.NewButton(0, 0, 36, 32, "5x", func() {
		tc.Paused = false
		tc.SetSpeed(5.0)
	})
	timeControlContainer.Add(speed5Btn)

	// 25x speed
	speed25Btn := widget.NewButton(0, 0, 42, 32, "25x", func() {
		tc.Paused = false
		tc.SetSpeed(25.0)
	})
	timeControlContainer.Add(speed25Btn)

	topBarContainer.Add(timeControlContainer)

	// Add topbar to root
	root.Add(topBar)
	root.Add(topBarContainer)

	// Position middle buttons and time control (will be updated in Update)
	g.repositionTopBarElements(800, 600) // Default size
}

func (g *Game) repositionTopBarElements(screenW, screenH float64) {
	root := g.UIManager.Root()
	if root.ChildCount() < 2 {
		return // Not fully initialized
	}

	topBarHeight := g.Config.UI.TopBarHeight

	// Update topbar panel size (first child)
	if panel := root.GetChild(0); panel != nil {
		panel.SetBounds(widget.Bounds{X: 0, Y: 0, W: screenW, H: topBarHeight})
	}

	// Update topbar container (second child)
	if topBarContainer, ok := root.GetChild(1).(*widget.Container); ok && topBarContainer != nil {
		topBarContainer.SetBounds(widget.Bounds{X: 0, Y: 0, W: screenW, H: topBarHeight})

		// Position middle buttons container in the center (third child of topBarContainer)
		if middleContainer, ok := topBarContainer.GetChild(2).(*widget.Container); ok && middleContainer != nil {
			middleW := 300.0
			middleContainer.SetBounds(widget.Bounds{
				X: (screenW - middleW) / 2,
				Y: 0,
				W: middleW,
				H: topBarHeight,
			})
		}

		// Position time control on the right (fourth child of topBarContainer)
		if timeContainer, ok := topBarContainer.GetChild(3).(*widget.Container); ok && timeContainer != nil {
			timeW := 180.0
			timeContainer.SetBounds(widget.Bounds{
				X: screenW - timeW - g.UIManager.Theme().PaddingLarge,
				Y: 0,
				W: timeW,
				H: topBarHeight,
			})
		}
	}
}

// Helper function (not used in final version since we rebuild in New)
func buildTopBar(uiMgr *ui.Manager, cfg *config.Config, tc *timecontrol.TimeControl) {
	// Placeholder - actual building happens in Game.buildTopBar()
}

func (g *Game) Update() error {
	sw, sh := g.Map.ScreenSize()

	// Update FPS label
	if g.fpsLabel != nil {
		g.fpsLabel.SetText(fmt.Sprintf("%.0f FPS", ebiten.ActualFPS()))
	}

	// Reposition topbar elements based on screen size
	g.repositionTopBarElements(sw, sh)

	consoleFocused := g.Console.Update(sw, sh)

	uiFocused := false
	if !consoleFocused {
		g.UIManager.Update(sw, sh)
		uiFocused = g.OldPanels.Update(sw, sh)
	}

	g.Map.SetInputSuppressed(uiFocused || consoleFocused)

	g.Map.UpdateInput()
	g.Time.Update(1.0 / 60.0)
	g.Map.UpdateLayers(g.Time)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.Map.Draw(screen)
	g.UIManager.Draw(screen)
	g.OldPanels.Draw(screen)
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
