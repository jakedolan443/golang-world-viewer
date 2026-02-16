
package ui

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Anchor int

const (
	AnchorTopLeft Anchor = iota
	AnchorTopCenter
	AnchorTopRight
	AnchorBottomLeft
	AnchorBottomCenter
	AnchorBottomRight
	AnchorTopStretch 
)

type Style struct {
	Background   color.RGBA
	Border       color.RGBA
	BorderWidth  float32
	CornerRadius float32
}

func DefaultPanelStyle() Style {
	return Style{
		Background:   color.RGBA{20, 24, 30, 210},
		Border:       color.RGBA{60, 68, 80, 180},
		BorderWidth:  1.0,
		CornerRadius: 6.0,
	}
}

func DefaultTopbarStyle() Style {
	return Style{
		Background:   color.RGBA{16, 19, 25, 230},
		Border:       color.RGBA{50, 58, 70, 200},
		BorderWidth:  1.0,
		CornerRadius: 0,
	}
}

type DrawContext struct {
	Screen *ebiten.Image
	X, Y   float64 
	W, H   float64 
}

type ContentDrawFunc func(ctx *DrawContext)
type ClickFunc func(ctx *DrawContext, mx, my float64)
type UpdateFunc func(ctx *DrawContext)

type Panel struct {
	ID       string
	Anchor   Anchor
	OffsetX  float64
	OffsetY  float64
	Width    float64
	Height   float64
	Padding  float64
	Style    Style
	OnDraw   ContentDrawFunc
	OnClick  ClickFunc
	OnUpdate UpdateFunc
	Visible  bool
	Hovered  bool
	Clicked  bool

	screenX, screenY float64
	screenW, screenH float64
}

func (p *Panel) ContentRect() (float64, float64, float64, float64) {
	pad := p.Padding
	return p.screenX + pad, p.screenY + pad, p.screenW - 2*pad, p.screenH - 2*pad
}

type Layer struct {
	panels []*Panel
}

func NewLayer() *Layer {
	return &Layer{}
}

func (u *Layer) AddPanel(p *Panel) {
	u.panels = append(u.panels, p)
}

func (u *Layer) Panel(id string) *Panel {
	for _, p := range u.panels {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (u *Layer) Update(screenW, screenH float64) bool {
	mx, my := ebiten.CursorPosition()
	mxf, myf := float64(mx), float64(my)
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

	anyHovered := false
	for _, p := range u.panels {
		if !p.Visible {
			p.Hovered = false
			p.Clicked = false
			continue
		}
		u.resolveLayout(p, screenW, screenH)
		p.Hovered = mxf >= p.screenX && mxf <= p.screenX+p.screenW &&
			myf >= p.screenY && myf <= p.screenY+p.screenH
		p.Clicked = p.Hovered && clicked
		if p.Hovered {
			anyHovered = true
		}
		if p.OnUpdate != nil {
			pad := p.Padding
			p.OnUpdate(&DrawContext{
				X: p.screenX + pad, Y: p.screenY + pad,
				W: p.screenW - 2*pad, H: p.screenH - 2*pad,
			})
		}
		if p.Clicked && p.OnClick != nil {
			pad := p.Padding
			p.OnClick(&DrawContext{
				X: p.screenX + pad, Y: p.screenY + pad,
				W: p.screenW - 2*pad, H: p.screenH - 2*pad,
			}, mxf, myf)
		}
	}
	return anyHovered
}

func (u *Layer) Draw(screen *ebiten.Image) {
	for _, p := range u.panels {
		if !p.Visible {
			continue
		}
		u.drawPanel(screen, p)
	}
}

func (u *Layer) resolveLayout(p *Panel, sw, sh float64) {
	w, h := p.Width, p.Height

	switch p.Anchor {
	case AnchorTopStretch:
		p.screenX = p.OffsetX
		p.screenY = p.OffsetY
		p.screenW = sw - 2*p.OffsetX
		p.screenH = h
		return
	case AnchorTopLeft:
		p.screenX = p.OffsetX
		p.screenY = p.OffsetY
	case AnchorTopCenter:
		p.screenX = (sw-w)/2 + p.OffsetX
		p.screenY = p.OffsetY
	case AnchorTopRight:
		p.screenX = sw - w - p.OffsetX
		p.screenY = p.OffsetY
	case AnchorBottomLeft:
		p.screenX = p.OffsetX
		p.screenY = sh - h - p.OffsetY
	case AnchorBottomCenter:
		p.screenX = (sw-w)/2 + p.OffsetX
		p.screenY = sh - h - p.OffsetY
	case AnchorBottomRight:
		p.screenX = sw - w - p.OffsetX
		p.screenY = sh - h - p.OffsetY
	}

	p.screenW = w
	p.screenH = h
}

func (u *Layer) drawPanel(screen *ebiten.Image, p *Panel) {
	x := float32(p.screenX)
	y := float32(p.screenY)
	w := float32(p.screenW)
	h := float32(p.screenH)
	st := p.Style

	if st.Background.A > 0 {
		if st.CornerRadius > 0 {
			drawRoundedRectFill(screen, x, y, w, h, st.CornerRadius, st.Background)
		} else {
			vector.DrawFilledRect(screen, x, y, w, h, st.Background, true)
		}
	}

	if st.Border.A > 0 && st.BorderWidth > 0 {
		if st.CornerRadius > 0 {
			strokeRoundedRect(screen, x, y, w, h, st.CornerRadius, st.BorderWidth, st.Border)
		} else {
			strokeRect(screen, x, y, w, h, st.BorderWidth, st.Border)
		}
	}

	if p.OnDraw != nil {
		pad := p.Padding
		p.OnDraw(&DrawContext{
			Screen: screen,
			X:      p.screenX + pad,
			Y:      p.screenY + pad,
			W:      p.screenW - 2*pad,
			H:      p.screenH - 2*pad,
		})
	}
}

func drawRoundedRectFill(dst *ebiten.Image, x, y, w, h, r float32, clr color.RGBA) {
	if r > w/2 {
		r = w / 2
	}
	if r > h/2 {
		r = h / 2
	}
	vector.DrawFilledRect(dst, x+r, y, w-2*r, h, clr, true)
	vector.DrawFilledRect(dst, x, y+r, w, h-2*r, clr, true)
	vector.DrawFilledCircle(dst, x+r, y+r, r, clr, true)
	vector.DrawFilledCircle(dst, x+w-r, y+r, r, clr, true)
	vector.DrawFilledCircle(dst, x+r, y+h-r, r, clr, true)
	vector.DrawFilledCircle(dst, x+w-r, y+h-r, r, clr, true)
}

func strokeRoundedRect(dst *ebiten.Image, x, y, w, h, r, bw float32, clr color.RGBA) {
	if r > w/2 {
		r = w / 2
	}
	if r > h/2 {
		r = h / 2
	}
	vector.StrokeLine(dst, x+r, y, x+w-r, y, bw, clr, true)
	vector.StrokeLine(dst, x+r, y+h, x+w-r, y+h, bw, clr, true)
	vector.StrokeLine(dst, x, y+r, x, y+h-r, bw, clr, true)
	vector.StrokeLine(dst, x+w, y+r, x+w, y+h-r, bw, clr, true)

	const segs = 8
	strokeArc(dst, x+r, y+r, r, math.Pi, 3*math.Pi/2, segs, bw, clr)
	strokeArc(dst, x+w-r, y+r, r, 3*math.Pi/2, 2*math.Pi, segs, bw, clr)
	strokeArc(dst, x+r, y+h-r, r, math.Pi/2, math.Pi, segs, bw, clr)
	strokeArc(dst, x+w-r, y+h-r, r, 0, math.Pi/2, segs, bw, clr)
}

func strokeArc(dst *ebiten.Image, cx, cy, r float32, a0, a1 float64, segs int, bw float32, clr color.RGBA) {
	step := (a1 - a0) / float64(segs)
	for i := 0; i < segs; i++ {
		t0 := a0 + step*float64(i)
		t1 := t0 + step
		x0 := cx + r*float32(math.Cos(t0))
		y0 := cy + r*float32(math.Sin(t0))
		x1 := cx + r*float32(math.Cos(t1))
		y1 := cy + r*float32(math.Sin(t1))
		vector.StrokeLine(dst, x0, y0, x1, y1, bw, clr, true)
	}
}

func strokeRect(dst *ebiten.Image, x, y, w, h, bw float32, clr color.RGBA) {
	vector.StrokeLine(dst, x, y, x+w, y, bw, clr, true)
	vector.StrokeLine(dst, x+w, y, x+w, y+h, bw, clr, true)
	vector.StrokeLine(dst, x+w, y+h, x, y+h, bw, clr, true)
	vector.StrokeLine(dst, x, y+h, x, y, bw, clr, true)
}
