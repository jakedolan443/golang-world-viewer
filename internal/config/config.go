
package config

import (
	"image/color"
	"strings"
)

type Quality int

const (
	QualityLow Quality = iota
	QualityMedium
	QualityHigh
	QualityUltra
)

func (q Quality) String() string {
	switch q {
	case QualityLow:
		return "Low"
	case QualityMedium:
		return "Medium"
	case QualityHigh:
		return "High"
	case QualityUltra:
		return "Ultra"
	default:
		return "Unknown"
	}
}

type WindowMode int

const (
	WindowModeWindowed WindowMode = iota
	WindowModeFullscreen
	WindowModeBorderless
)

func (m WindowMode) String() string {
	switch m {
	case WindowModeWindowed:
		return "Windowed"
	case WindowModeFullscreen:
		return "Fullscreen"
	case WindowModeBorderless:
		return "Borderless"
	default:
		return "Unknown"
	}
}

type Resolution struct {
	Name   string
	Width  int
	Height int
}

var Resolutions = []Resolution{
	{"480p", 854, 480},
	{"720p", 1280, 720},
	{"900p", 1600, 900},
	{"1080p", 1920, 1080},
	{"1440p", 2560, 1440},
	{"4k", 3840, 2160},
}

func ResolutionByName(name string) (Resolution, bool) {
	lower := strings.ToLower(name)
	for _, r := range Resolutions {
		if strings.ToLower(r.Name) == lower {
			return r, true
		}
	}
	return Resolution{}, false
}

func ResolutionNames() []string {
	out := make([]string, len(Resolutions))
	for i, r := range Resolutions {
		out[i] = r.Name
	}
	return out
}

type DisplayConfig struct {
	WindowWidth  int
	WindowHeight int
	Resolution   Resolution
	WindowMode   WindowMode
	Monitor      int
	VSync        bool
	Resizable    bool
	FilterLinear bool
}

type MapConfig struct {
	ShapefilePath  string
	MaxZoom        float64
	MinZoomPadding float64 
	SmoothFactor   float64
	ShowEquator    bool
	ShowDebug      bool

	ColorOcean      color.RGBA
	ColorLand       color.RGBA
	ColorAntarctica color.RGBA
	ColorEquator    color.RGBA
}

type CloudConfig struct {
	ZoomThreshold    float64
	FadeDuration     float64
	CellWorld        float64
	TargetPxPerSample float64
	MaxDimension     int
	WorkerCount      int 
}

type UIConfig struct {
	TopBarHeight    float64
	PanelPadding    float64
	BackgroundColor color.RGBA
	BorderColor     color.RGBA
	BorderWidth     float32
	CornerRadius    float32
}

type TimeConfig struct {
	DefaultSpeed float64
	SpeedOptions []float64
}

type Config struct {
	Title   string
	Quality Quality

	Display DisplayConfig
	Map     MapConfig
	Cloud   CloudConfig
	UI      UIConfig
	Time    TimeConfig
}

func Default() *Config {
	cfg := &Config{
		Title:   "Shipping Management",
		Quality: QualityHigh,

		Display: DisplayConfig{
			WindowWidth:  1280,
			WindowHeight: 720,
			Resolution:   Resolution{"1080p", 1920, 1080},
			WindowMode:   WindowModeWindowed,
			Monitor:      0,
			VSync:        true,
			Resizable:    true,
			FilterLinear: true,
		},

		Map: MapConfig{
			ShapefilePath:  "ne_50m_land/ne_50m_land.shp",
			MaxZoom:        80.0,
			MinZoomPadding: 0.95,
			SmoothFactor:   0.15,
			ShowEquator:    true,
			ShowDebug:      true,

			ColorOcean:      color.RGBA{42, 50, 62, 255},
			ColorLand:       color.RGBA{105, 110, 115, 255},
			ColorAntarctica: color.RGBA{165, 170, 175, 255},
			ColorEquator:    color.RGBA{58, 65, 78, 80},
		},

		Cloud: CloudConfig{
			ZoomThreshold:    30.0,
			FadeDuration:     1.0,
			CellWorld:        0.8,
			TargetPxPerSample: 8.0,
			MaxDimension:     800,
			WorkerCount:      0,
		},

		UI: UIConfig{
			TopBarHeight:    48,
			PanelPadding:    14,
			BackgroundColor: color.RGBA{20, 24, 30, 210},
			BorderColor:     color.RGBA{60, 68, 80, 180},
			BorderWidth:     1.0,
			CornerRadius:    6.0,
		},

		Time: TimeConfig{
			DefaultSpeed: 1.0,
			SpeedOptions: []float64{1.0, 5.0, 25.0},
		},
	}
	return cfg
}

func (c *Config) WithQuality(q Quality) *Config {
	cp := *c
	cp.Quality = q

	switch q {
	case QualityLow:
		cp.Cloud.TargetPxPerSample = 16.0
		cp.Cloud.MaxDimension = 400
		cp.Display.Resolution = Resolution{"720p", 1280, 720}
		cp.Display.FilterLinear = false
	case QualityMedium:
		cp.Cloud.TargetPxPerSample = 12.0
		cp.Cloud.MaxDimension = 600
		cp.Display.Resolution = Resolution{"900p", 1600, 900}
		cp.Display.FilterLinear = true
	case QualityHigh:
		cp.Cloud.TargetPxPerSample = 8.0
		cp.Cloud.MaxDimension = 800
		cp.Display.Resolution = Resolution{"1080p", 1920, 1080}
		cp.Display.FilterLinear = true
	case QualityUltra:
		cp.Cloud.TargetPxPerSample = 4.0
		cp.Cloud.MaxDimension = 1200
		cp.Display.Resolution = Resolution{"1440p", 2560, 1440}
		cp.Display.FilterLinear = true
	}

	return &cp
}
