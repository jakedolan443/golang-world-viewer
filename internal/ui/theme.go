package ui

import "image/color"

// Theme defines the visual appearance of UI elements.
type Theme struct {
	// Button colors
	ButtonNormal      color.RGBA
	ButtonHover       color.RGBA
	ButtonPressed     color.RGBA
	ButtonText        color.RGBA
	ButtonBorder      color.RGBA
	ButtonBorderWidth float32
	ButtonRadius      float32

	// Panel colors
	PanelBackground   color.RGBA
	PanelBorder       color.RGBA
	PanelBorderWidth  float32
	PanelRadius       float32

	// Text colors
	TextPrimary   color.RGBA
	TextSecondary color.RGBA
	TextDisabled  color.RGBA

	// Padding and spacing
	PaddingSmall  float64
	PaddingMedium float64
	PaddingLarge  float64
	Spacing       float64
}

// DefaultTheme returns the default theme for the application.
func DefaultTheme() *Theme {
	return &Theme{
		// Button colors - neutral gray
		ButtonNormal:      color.RGBA{45, 52, 62, 220},
		ButtonHover:       color.RGBA{60, 70, 82, 230},
		ButtonPressed:     color.RGBA{35, 42, 50, 240},
		ButtonText:        color.RGBA{220, 220, 220, 255},
		ButtonBorder:      color.RGBA{80, 90, 105, 200},
		ButtonBorderWidth: 1.0,
		ButtonRadius:      4.0,

		// Panel colors
		PanelBackground:   color.RGBA{20, 24, 30, 210},
		PanelBorder:       color.RGBA{60, 68, 80, 180},
		PanelBorderWidth:  1.0,
		PanelRadius:       6.0,

		// Text colors
		TextPrimary:   color.RGBA{220, 220, 220, 255},
		TextSecondary: color.RGBA{160, 160, 160, 255},
		TextDisabled:  color.RGBA{100, 100, 100, 255},

		// Spacing
		PaddingSmall:  4.0,
		PaddingMedium: 8.0,
		PaddingLarge:  12.0,
		Spacing:       4.0,
	}
}

// ColoredButtonTheme creates a button color scheme.
func ColoredButtonTheme(baseColor color.RGBA) (normal, hover, pressed color.RGBA) {
	// Normal: base color
	normal = baseColor

	// Hover: slightly lighter
	hover = color.RGBA{
		R: min(255, baseColor.R+20),
		G: min(255, baseColor.G+20),
		B: min(255, baseColor.B+20),
		A: baseColor.A,
	}

	// Pressed: darker
	pressed = color.RGBA{
		R: baseColor.R * 85 / 100,
		G: baseColor.G * 85 / 100,
		B: baseColor.B * 85 / 100,
		A: baseColor.A,
	}

	return
}

func min(a, b uint8) uint8 {
	if a < b {
		return a
	}
	return b
}
