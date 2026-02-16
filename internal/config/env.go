package config

import (
	"bufio"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"
)

func LoadEnv(path string) (*Config, error) {
	cfg := Default()

	values, err := parseEnvFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading env file: %w", err)
	}

	applyEnv(cfg, values)
	return cfg, nil
}

func parseEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		val = strings.Trim(val, `"'`)
		values[strings.ToUpper(key)] = val
	}
	return values, scanner.Err()
}

func applyEnv(cfg *Config, v map[string]string) {
	setString(v, "TITLE", &cfg.Title)

	if q, ok := v["QUALITY"]; ok {
		switch strings.ToLower(q) {
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

	setInt(v, "WINDOW_WIDTH", &cfg.Display.WindowWidth)
	setInt(v, "WINDOW_HEIGHT", &cfg.Display.WindowHeight)
	setResolution(v, "RESOLUTION", &cfg.Display.Resolution)
	setWindowMode(v, "WINDOW_MODE", &cfg.Display.WindowMode)
	setInt(v, "MONITOR", &cfg.Display.Monitor)
	setBool(v, "VSYNC", &cfg.Display.VSync)
	setBool(v, "RESIZABLE", &cfg.Display.Resizable)
	setBool(v, "FILTER_LINEAR", &cfg.Display.FilterLinear)

	setString(v, "SHAPEFILE_PATH", &cfg.Map.ShapefilePath)
	setFloat64(v, "MAX_ZOOM", &cfg.Map.MaxZoom)
	setFloat64(v, "MIN_ZOOM_PADDING", &cfg.Map.MinZoomPadding)
	setFloat64(v, "SMOOTH_FACTOR", &cfg.Map.SmoothFactor)
	setBool(v, "SHOW_EQUATOR", &cfg.Map.ShowEquator)
	setBool(v, "SHOW_DEBUG", &cfg.Map.ShowDebug)
	setColor(v, "COLOR_OCEAN", &cfg.Map.ColorOcean)
	setColor(v, "COLOR_LAND", &cfg.Map.ColorLand)
	setColor(v, "COLOR_ANTARCTICA", &cfg.Map.ColorAntarctica)
	setColor(v, "COLOR_EQUATOR", &cfg.Map.ColorEquator)

	setFloat64(v, "CLOUD_ZOOM_THRESHOLD", &cfg.Cloud.ZoomThreshold)
	setFloat64(v, "CLOUD_FADE_DURATION", &cfg.Cloud.FadeDuration)
	setFloat64(v, "CLOUD_CELL_WORLD", &cfg.Cloud.CellWorld)
	setFloat64(v, "CLOUD_TARGET_PX", &cfg.Cloud.TargetPxPerSample)
	setInt(v, "CLOUD_MAX_DIM", &cfg.Cloud.MaxDimension)
	setInt(v, "CLOUD_WORKERS", &cfg.Cloud.WorkerCount)

	setFloat64(v, "UI_TOPBAR_HEIGHT", &cfg.UI.TopBarHeight)
	setFloat64(v, "UI_PANEL_PADDING", &cfg.UI.PanelPadding)
	setColor(v, "UI_BACKGROUND", &cfg.UI.BackgroundColor)
	setColor(v, "UI_BORDER", &cfg.UI.BorderColor)
	setFloat32(v, "UI_BORDER_WIDTH", &cfg.UI.BorderWidth)
	setFloat32(v, "UI_CORNER_RADIUS", &cfg.UI.CornerRadius)

	setFloat64(v, "TIME_DEFAULT_SPEED", &cfg.Time.DefaultSpeed)
}

func setString(v map[string]string, key string, dst *string) {
	if s, ok := v[key]; ok {
		*dst = s
	}
}

func setInt(v map[string]string, key string, dst *int) {
	if s, ok := v[key]; ok {
		if n, err := strconv.Atoi(s); err == nil {
			*dst = n
		}
	}
}

func setBool(v map[string]string, key string, dst *bool) {
	if s, ok := v[key]; ok {
		switch strings.ToLower(s) {
		case "true", "1", "yes":
			*dst = true
		case "false", "0", "no":
			*dst = false
		}
	}
}

func setFloat64(v map[string]string, key string, dst *float64) {
	if s, ok := v[key]; ok {
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			*dst = f
		}
	}
}

func setFloat32(v map[string]string, key string, dst *float32) {
	if s, ok := v[key]; ok {
		if f, err := strconv.ParseFloat(s, 32); err == nil {
			*dst = float32(f)
		}
	}
}

func setWindowMode(v map[string]string, key string, dst *WindowMode) {
	if s, ok := v[key]; ok {
		switch strings.ToLower(s) {
		case "windowed", "window":
			*dst = WindowModeWindowed
		case "fullscreen", "full":
			*dst = WindowModeFullscreen
		case "borderless", "borderless_fullscreen":
			*dst = WindowModeBorderless
		}
	}
}

func setResolution(v map[string]string, key string, dst *Resolution) {
	if s, ok := v[key]; ok {
		if r, found := ResolutionByName(s); found {
			*dst = r
		}
	}
}

func setColor(v map[string]string, key string, dst *color.RGBA) {
	s, ok := v[key]
	if !ok {
		return
	}
	parts := strings.Split(s, ",")
	if len(parts) < 3 || len(parts) > 4 {
		return
	}
	vals := make([]uint8, len(parts))
	for i, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || n < 0 || n > 255 {
			return
		}
		vals[i] = uint8(n)
	}
	dst.R = vals[0]
	dst.G = vals[1]
	dst.B = vals[2]
	if len(vals) == 4 {
		dst.A = vals[3]
	}
}
