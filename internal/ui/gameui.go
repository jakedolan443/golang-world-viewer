package ui

import (
	"bytes"
	"image/color"
	"log"

	"github.com/ebitenui/ebitenui"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/goregular"
)

var (
	fontFace     text.Face
	fontFaceBold text.Face
)

func init() {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		log.Fatalf("failed to load font: %v", err)
	}
	fontFace = &text.GoTextFace{Source: src, Size: 14}
	fontFaceBold = &text.GoTextFace{Source: src, Size: 14}
}

// Face returns the default font face.
func Face() text.Face { return fontFace }

// FaceSize returns a font face at the given size.
func FaceSize(size float64) text.Face {
	src, _ := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	return &text.GoTextFace{Source: src, Size: size}
}

// NewUI creates a fresh ebitenui.UI with the given root container.
func NewUI(root *widget.Container) *ebitenui.UI {
	return &ebitenui.UI{Container: root}
}

// ----- Button helpers -----

// ButtonImage builds a ButtonImage from three NRGBA colors.
func ButtonImage(idle, hover, pressed color.NRGBA) *widget.ButtonImage {
	return &widget.ButtonImage{
		Idle:    euiimage.NewNineSliceColor(idle),
		Hover:   euiimage.NewNineSliceColor(hover),
		Pressed: euiimage.NewNineSliceColor(pressed),
	}
}

// ButtonColors generates idle/hover/pressed from a base color.
func ButtonColors(base color.NRGBA) *widget.ButtonImage {
	hover := color.NRGBA{
		R: clampAdd(base.R, 20),
		G: clampAdd(base.G, 20),
		B: clampAdd(base.B, 20),
		A: base.A,
	}
	pressed := color.NRGBA{
		R: base.R * 80 / 100,
		G: base.G * 80 / 100,
		B: base.B * 80 / 100,
		A: base.A,
	}
	return ButtonImage(base, hover, pressed)
}

// TextButton creates a button with text label and click handler.
func TextButton(label string, img *widget.ButtonImage, textColor color.Color, face text.Face, onClick func()) *widget.Button {
	opts := []widget.ButtonOpt{
		widget.ButtonOpts.Image(img),
		widget.ButtonOpts.Text(label, &face, &widget.ButtonTextColor{
			Idle:    textColor,
			Hover:   textColor,
			Pressed: textColor,
		}),
		widget.ButtonOpts.TextPadding(&widget.Insets{
			Left: 12, Right: 12, Top: 4, Bottom: 4,
		}),
	}
	if onClick != nil {
		opts = append(opts, widget.ButtonOpts.ClickedHandler(
			func(args *widget.ButtonClickedEventArgs) { onClick() },
		))
	}
	return widget.NewButton(opts...)
}

// ----- Container helpers -----

// HRow creates a horizontal row container with the given spacing.
func HRow(spacing int, children ...widget.PreferredSizeLocateableWidget) *widget.Container {
	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionHorizontal),
			widget.RowLayoutOpts.Spacing(spacing),
		)),
	)
	for _, child := range children {
		c.AddChild(child)
	}
	return c
}

// VCol creates a vertical column container with the given spacing.
func VCol(spacing int, children ...widget.PreferredSizeLocateableWidget) *widget.Container {
	c := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(spacing),
		)),
	)
	for _, child := range children {
		c.AddChild(child)
	}
	return c
}

// ----- Label helper -----

// Label creates a text label.
func Label(t string, face text.Face, clr color.Color) *widget.Text {
	return widget.NewText(
		widget.TextOpts.Text(t, &face, clr),
	)
}

// ----- Panel helper -----

// PanelContainer creates a container with a background color.
func PanelContainer(bg color.NRGBA, opts ...widget.ContainerOpt) *widget.Container {
	all := []widget.ContainerOpt{
		widget.ContainerOpts.BackgroundImage(euiimage.NewNineSliceColor(bg)),
	}
	all = append(all, opts...)
	return widget.NewContainer(all...)
}

func clampAdd(v uint8, delta uint8) uint8 {
	if int(v)+int(delta) > 255 {
		return 255
	}
	return v + delta
}
