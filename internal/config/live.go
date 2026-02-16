package config

import (
	"fmt"
	"image/color"
	"sort"
	"strings"
)

var keyList = []string{
	"TITLE",
	"QUALITY",
	"RESOLUTION",
	"WINDOW_WIDTH", "WINDOW_HEIGHT",
	"WINDOW_MODE", "MONITOR",
	"VSYNC", "RESIZABLE", "FILTER_LINEAR",
	"MAX_ZOOM", "MIN_ZOOM_PADDING", "SMOOTH_FACTOR",
	"SHOW_EQUATOR", "SHOW_DEBUG",
	"COLOR_OCEAN", "COLOR_LAND", "COLOR_ANTARCTICA", "COLOR_EQUATOR",
	"CLOUD_ZOOM_THRESHOLD", "CLOUD_FADE_DURATION",
	"CLOUD_CELL_WORLD", "CLOUD_TARGET_PX",
	"CLOUD_MAX_DIM",
	"UI_TOPBAR_HEIGHT", "UI_PANEL_PADDING",
	"UI_BACKGROUND", "UI_BORDER",
	"UI_BORDER_WIDTH", "UI_CORNER_RADIUS",
	"TIME_DEFAULT_SPEED",
}

var initOnlyKeys = map[string]bool{
	"SHAPEFILE_PATH": true,
	"CLOUD_WORKERS":  true,
}

func AllKeys() []string {
	out := make([]string, len(keyList))
	copy(out, keyList)
	sort.Strings(out)
	return out
}

func GetValue(cfg *Config, key string) string {
	switch strings.ToUpper(key) {
	case "TITLE":
		return cfg.Title
	case "QUALITY":
		return cfg.Quality.String()
	case "WINDOW_WIDTH":
		return fmt.Sprintf("%d", cfg.Display.WindowWidth)
	case "WINDOW_HEIGHT":
		return fmt.Sprintf("%d", cfg.Display.WindowHeight)
	case "RESOLUTION":
		return fmt.Sprintf("%s (%dx%d)", cfg.Display.Resolution.Name,
			cfg.Display.Resolution.Width, cfg.Display.Resolution.Height)
	case "WINDOW_MODE":
		return cfg.Display.WindowMode.String()
	case "MONITOR":
		return fmt.Sprintf("%d", cfg.Display.Monitor)
	case "VSYNC":
		return fmt.Sprintf("%t", cfg.Display.VSync)
	case "RESIZABLE":
		return fmt.Sprintf("%t", cfg.Display.Resizable)
	case "FILTER_LINEAR":
		return fmt.Sprintf("%t", cfg.Display.FilterLinear)
	case "SHAPEFILE_PATH":
		return cfg.Map.ShapefilePath
	case "MAX_ZOOM":
		return fmt.Sprintf("%.2f", cfg.Map.MaxZoom)
	case "MIN_ZOOM_PADDING":
		return fmt.Sprintf("%.2f", cfg.Map.MinZoomPadding)
	case "SMOOTH_FACTOR":
		return fmt.Sprintf("%.2f", cfg.Map.SmoothFactor)
	case "SHOW_EQUATOR":
		return fmt.Sprintf("%t", cfg.Map.ShowEquator)
	case "SHOW_DEBUG":
		return fmt.Sprintf("%t", cfg.Map.ShowDebug)
	case "COLOR_OCEAN":
		return fmtColor(cfg.Map.ColorOcean)
	case "COLOR_LAND":
		return fmtColor(cfg.Map.ColorLand)
	case "COLOR_ANTARCTICA":
		return fmtColor(cfg.Map.ColorAntarctica)
	case "COLOR_EQUATOR":
		return fmtColor(cfg.Map.ColorEquator)
	case "CLOUD_ZOOM_THRESHOLD":
		return fmt.Sprintf("%.1f", cfg.Cloud.ZoomThreshold)
	case "CLOUD_FADE_DURATION":
		return fmt.Sprintf("%.1f", cfg.Cloud.FadeDuration)
	case "CLOUD_CELL_WORLD":
		return fmt.Sprintf("%.1f", cfg.Cloud.CellWorld)
	case "CLOUD_TARGET_PX":
		return fmt.Sprintf("%.1f", cfg.Cloud.TargetPxPerSample)
	case "CLOUD_MAX_DIM":
		return fmt.Sprintf("%d", cfg.Cloud.MaxDimension)
	case "CLOUD_WORKERS":
		return fmt.Sprintf("%d", cfg.Cloud.WorkerCount)
	case "UI_TOPBAR_HEIGHT":
		return fmt.Sprintf("%.0f", cfg.UI.TopBarHeight)
	case "UI_PANEL_PADDING":
		return fmt.Sprintf("%.0f", cfg.UI.PanelPadding)
	case "UI_BACKGROUND":
		return fmtColor(cfg.UI.BackgroundColor)
	case "UI_BORDER":
		return fmtColor(cfg.UI.BorderColor)
	case "UI_BORDER_WIDTH":
		return fmt.Sprintf("%.1f", cfg.UI.BorderWidth)
	case "UI_CORNER_RADIUS":
		return fmt.Sprintf("%.1f", cfg.UI.CornerRadius)
	case "TIME_DEFAULT_SPEED":
		return fmt.Sprintf("%.1f", cfg.Time.DefaultSpeed)
	default:
		return ""
	}
}

func SetValue(cfg *Config, key, value string) string {
	k := strings.ToUpper(strings.TrimSpace(key))
	value = strings.TrimSpace(value)

	if initOnlyKeys[k] {
		return fmt.Sprintf("%s can only be set in .env (requires restart)", k)
	}

	before := GetValue(cfg, k)
	if before == "" {
		return fmt.Sprintf("unknown key: %s", k)
	}

	v := map[string]string{k: value}

	switch k {
	case "TITLE":
		setString(v, k, &cfg.Title)
	case "QUALITY":
		applyQuality(cfg, value)
	case "RESOLUTION":
		if r, found := ResolutionByName(value); found {
			cfg.Display.Resolution = r
		} else {
			return fmt.Sprintf("unknown resolution: %s (try: %s)",
				value, strings.Join(ResolutionNames(), ", "))
		}
	case "WINDOW_WIDTH":
		setInt(v, k, &cfg.Display.WindowWidth)
	case "WINDOW_HEIGHT":
		setInt(v, k, &cfg.Display.WindowHeight)
	case "WINDOW_MODE":
		setWindowMode(v, k, &cfg.Display.WindowMode)
	case "MONITOR":
		setInt(v, k, &cfg.Display.Monitor)
	case "VSYNC":
		setBool(v, k, &cfg.Display.VSync)
	case "RESIZABLE":
		setBool(v, k, &cfg.Display.Resizable)
	case "FILTER_LINEAR":
		setBool(v, k, &cfg.Display.FilterLinear)
	case "SHAPEFILE_PATH":
		setString(v, k, &cfg.Map.ShapefilePath)
	case "MAX_ZOOM":
		setFloat64(v, k, &cfg.Map.MaxZoom)
	case "MIN_ZOOM_PADDING":
		setFloat64(v, k, &cfg.Map.MinZoomPadding)
	case "SMOOTH_FACTOR":
		setFloat64(v, k, &cfg.Map.SmoothFactor)
	case "SHOW_EQUATOR":
		setBool(v, k, &cfg.Map.ShowEquator)
	case "SHOW_DEBUG":
		setBool(v, k, &cfg.Map.ShowDebug)
	case "COLOR_OCEAN":
		setColor(v, k, &cfg.Map.ColorOcean)
	case "COLOR_LAND":
		setColor(v, k, &cfg.Map.ColorLand)
	case "COLOR_ANTARCTICA":
		setColor(v, k, &cfg.Map.ColorAntarctica)
	case "COLOR_EQUATOR":
		setColor(v, k, &cfg.Map.ColorEquator)
	case "CLOUD_ZOOM_THRESHOLD":
		setFloat64(v, k, &cfg.Cloud.ZoomThreshold)
	case "CLOUD_FADE_DURATION":
		setFloat64(v, k, &cfg.Cloud.FadeDuration)
	case "CLOUD_CELL_WORLD":
		setFloat64(v, k, &cfg.Cloud.CellWorld)
	case "CLOUD_TARGET_PX":
		setFloat64(v, k, &cfg.Cloud.TargetPxPerSample)
	case "CLOUD_MAX_DIM":
		setInt(v, k, &cfg.Cloud.MaxDimension)
	case "CLOUD_WORKERS":
		setInt(v, k, &cfg.Cloud.WorkerCount)
	case "UI_TOPBAR_HEIGHT":
		setFloat64(v, k, &cfg.UI.TopBarHeight)
	case "UI_PANEL_PADDING":
		setFloat64(v, k, &cfg.UI.PanelPadding)
	case "UI_BACKGROUND":
		setColor(v, k, &cfg.UI.BackgroundColor)
	case "UI_BORDER":
		setColor(v, k, &cfg.UI.BorderColor)
	case "UI_BORDER_WIDTH":
		setFloat32(v, k, &cfg.UI.BorderWidth)
	case "UI_CORNER_RADIUS":
		setFloat32(v, k, &cfg.UI.CornerRadius)
	case "TIME_DEFAULT_SPEED":
		setFloat64(v, k, &cfg.Time.DefaultSpeed)
	default:
		return fmt.Sprintf("unknown key: %s", k)
	}

	after := GetValue(cfg, k)
	if after == before {
		return fmt.Sprintf("%s unchanged (still %s) — bad value?", k, before)
	}
	return fmt.Sprintf("%s = %s (was %s)", k, after, before)
}

func applyQuality(cfg *Config, value string) {
	switch strings.ToLower(value) {
	case "low":
		*cfg = *cfg.WithQuality(QualityLow)
	case "medium":
		*cfg = *cfg.WithQuality(QualityMedium)
	case "high":
		*cfg = *cfg.WithQuality(QualityHigh)
	case "ultra":
		*cfg = *cfg.WithQuality(QualityUltra)
	}
}

func fmtColor(c color.RGBA) string {
	return fmt.Sprintf("%d,%d,%d,%d", c.R, c.G, c.B, c.A)
}

func ValueOptions(key string) []string {
	switch strings.ToUpper(key) {
	case "QUALITY":
		return []string{"Low", "Medium", "High", "Ultra"}
	case "RESOLUTION":
		return ResolutionNames()
	case "WINDOW_MODE":
		return []string{"Windowed", "Fullscreen", "Borderless"}
	case "VSYNC", "RESIZABLE", "FILTER_LINEAR", "SHOW_EQUATOR", "SHOW_DEBUG":
		return []string{"true", "false"}
	default:
		return nil
	}
}
