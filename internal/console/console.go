package console

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"shipping/internal/config"
)

const (
	maxOutputLines  = 256
	maxCmdHistory   = 64
	charW           = 6.0
	lineH           = 16.0
	inputPrefix     = "> "
	cursorBlinkRate = 30
	panelMinW       = 300
	panelMinH       = 180
	panelDefaultW   = 620
	panelDefaultH   = 360
	titleBarH       = 24.0
	pad             = 8.0
	scrollLines     = 3
	resizeGripSize  = 16.0
)

var (
	colBg      = color.RGBA{10, 12, 16, 230}
	colTitle   = color.RGBA{20, 30, 20, 240}
	colBorder  = color.RGBA{60, 180, 120, 200}
	colText    = color.RGBA{190, 200, 190, 255}
	colPrompt  = color.RGBA{80, 200, 120, 255}
	colHint    = color.RGBA{60, 75, 60, 160}
	colError   = color.RGBA{220, 80, 80, 255}
	colSelect  = color.RGBA{60, 120, 80, 120}
	colCursor  = color.RGBA{80, 200, 120, 255}
	colGripDot = color.RGBA{60, 180, 120, 120}
)

type line struct {
	text string
	clr  color.RGBA
}

type Console struct {
	Open bool
	cfg  *config.Config

	px, py, pw, ph float64

	dragging         bool
	dragOX, dragOY   float64
	resizing         bool
	resizeOX, resizeOY float64

	output []line
	input  []rune
	cursor int

	selAnchor int
	selCursor int
	selActive bool

	cmdHistory []string
	cmdIdx     int

	scrollOff  int
	blink      int
	suggestion string
	allKeys    []string
	justOpened bool
	OnSet      func(key string)

	tabOptions []string
	tabIndex   int
	tabBase    string
}

func New(cfg *config.Config) *Console {
	c := &Console{
		cfg:       cfg,
		px:        40,
		py:        40,
		pw:        panelDefaultW,
		ph:        panelDefaultH,
		selAnchor: -1,
		selCursor: -1,
		allKeys:   config.AllKeys(),
	}
	c.emit("Debug Console — type /help for commands", colPrompt)
	return c
}

func (c *Console) notifySet(key string) {
	if c.OnSet != nil {
		c.OnSet(key)
	}
}

func (c *Console) Update(screenW, screenH float64) bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyBackquote) {
		c.Open = !c.Open
		if c.Open {
			c.justOpened = true
			c.blink = 0
		}
		return c.Open
	}
	if !c.Open {
		return false
	}
	if c.justOpened {
		c.justOpened = false
		ebiten.AppendInputChars(nil)
		return true
	}

	mxi, myi := ebiten.CursorPosition()
	mx, my := float64(mxi), float64(myi)
	c.updateDragResize(mx, my, screenW, screenH)
	c.updateScroll(mx, my)
	c.blink++

	chars := ebiten.AppendInputChars(nil)
	for _, ch := range chars {
		if ch == '`' || ch == '~' {
			continue
		}
		c.deleteSelection()
		c.ins(ch)
	}

	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	shift := ebiten.IsKeyPressed(ebiten.KeyShift)

	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyC) {
		c.doCopy()
	}
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyV) {
		c.doPaste()
	}
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyA) {
		c.selAnchor = 0
		c.selCursor = len(c.input)
		c.selActive = true
		c.cursor = len(c.input)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) {
		c.submit()
	}

	c.repeatableKey(ebiten.KeyBackspace, func() { c.doBackspace() })
	c.repeatableKey(ebiten.KeyDelete, func() { c.doDelete() })

	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		c.completeTab()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		if shift {
			c.selExtend(-1)
		} else {
			c.selClear()
			if c.cursor > 0 {
				c.cursor--
			}
		}
		c.blink = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		if shift {
			c.selExtend(1)
		} else {
			c.selClear()
			if c.cursor < len(c.input) {
				c.cursor++
			}
		}
		c.blink = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyHome) {
		if shift {
			c.selExtendTo(0)
		} else {
			c.selClear()
		}
		c.cursor = 0
		c.blink = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnd) {
		if shift {
			c.selExtendTo(len(c.input))
		} else {
			c.selClear()
		}
		c.cursor = len(c.input)
		c.blink = 0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		c.historyUp()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		c.historyDown()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if c.hasSel() {
			c.selClear()
		} else if len(c.tabOptions) > 0 {
			c.tabOptions = nil
			c.tabBase = ""
			c.suggestion = ""
		} else if len(c.input) > 0 {
			c.input = c.input[:0]
			c.cursor = 0
			c.suggestion = ""
			c.tabOptions = nil
			c.tabBase = ""
		} else {
			c.Open = false
		}
	}

	c.updateSuggestion()

	return c.hitTest(mx, my) || c.dragging || c.resizing
}

func (c *Console) Draw(screen *ebiten.Image) {
	if !c.Open {
		return
	}

	x, y := float32(c.px), float32(c.py)
	w, h := float32(c.pw), float32(c.ph)
	tb := float32(titleBarH)

	vector.DrawFilledRect(screen, x, y, w, tb, colTitle, false)
	vector.DrawFilledRect(screen, x, y+tb, w, h-tb, colBg, false)
	vector.StrokeRect(screen, x, y, w, h, 1.5, colBorder, false)

	ebitenutil.DebugPrintAt(screen, "Console", int(x)+8, int(y)+5)
	ebitenutil.DebugPrintAt(screen, "[`]", int(x+w)-30, int(y)+5)

	cx := c.px + pad
	cy := c.py + titleBarH + pad
	ch := c.ph - titleBarH - pad*2 - lineH - 8

	visCount := int(ch / lineH)
	total := len(c.output)
	startIdx := total - visCount - c.scrollOff
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := startIdx + visCount
	if endIdx > total {
		endIdx = total
	}

	ly := cy
	for i := startIdx; i < endIdx; i++ {
		ebitenutil.DebugPrintAt(screen, c.output[i].text, int(cx), int(ly))
		ly += lineH
	}

	inputY := c.py + c.ph - lineH - pad
	prompt := inputPrefix + string(c.input)

	if c.hasSel() {
		s, e := c.selRange()
		sx := cx + float64(len(inputPrefix)+s)*charW
		ex := cx + float64(len(inputPrefix)+e)*charW
		vector.DrawFilledRect(screen,
			float32(sx), float32(inputY),
			float32(ex-sx), float32(lineH),
			colSelect, false)
	}

	ebitenutil.DebugPrintAt(screen, prompt, int(cx), int(inputY))

	if c.suggestion != "" {
		typed := string(c.input)
		if len(c.suggestion) > len(typed) {
			hintX := cx + float64(len(prompt))*charW
			ebitenutil.DebugPrintAt(screen, c.suggestion[len(typed):], int(hintX), int(inputY))
		}
	}

	if len(c.tabOptions) > 1 {
		dropY := inputY + lineH + 2
		dropX := cx
		maxW := float32(0)
		for _, opt := range c.tabOptions {
			ow := float32(len(opt)) * float32(charW) + 12
			if ow > maxW {
				maxW = ow
			}
		}
		dropH := float32(len(c.tabOptions)) * float32(lineH) + 4
		vector.DrawFilledRect(screen,
			float32(dropX)-2, float32(dropY)-2,
			maxW+4, dropH+2,
			color.RGBA{15, 20, 15, 240}, false)
		vector.StrokeRect(screen,
			float32(dropX)-2, float32(dropY)-2,
			maxW+4, dropH+2,
			1, colBorder, false)
		selIdx := c.tabIndex % len(c.tabOptions)
		for i, opt := range c.tabOptions {
			oy := dropY + float64(i)*lineH + 2
			if i == selIdx {
				vector.DrawFilledRect(screen,
					float32(dropX)-1, float32(oy),
					maxW+2, float32(lineH),
					color.RGBA{40, 70, 50, 200}, false)
				ebitenutil.DebugPrintAt(screen, opt, int(dropX+4), int(oy))
			} else {
				ebitenutil.DebugPrintAt(screen, opt, int(dropX+4), int(oy))
			}
		}
	}

	if (c.blink/cursorBlinkRate)%2 == 0 {
		curX := cx + float64(len(inputPrefix)+c.cursor)*charW
		vector.StrokeLine(screen,
			float32(curX), float32(inputY),
			float32(curX), float32(inputY+lineH-2),
			1, colCursor, false)
	}

	gx := x + w - 12
	gy := y + h - 12
	vector.StrokeLine(screen, gx+2, gy+10, gx+10, gy+2, 1, colGripDot, false)
	vector.StrokeLine(screen, gx+5, gy+10, gx+10, gy+5, 1, colGripDot, false)
	vector.StrokeLine(screen, gx+8, gy+10, gx+10, gy+8, 1, colGripDot, false)
}

func (c *Console) hitTest(mx, my float64) bool {
	return mx >= c.px && mx <= c.px+c.pw && my >= c.py && my <= c.py+c.ph
}

func (c *Console) visLines() int {
	ch := c.ph - titleBarH - pad*2 - lineH - 8
	if ch < 0 {
		return 0
	}
	return int(ch / lineH)
}

func (c *Console) updateDragResize(mx, my, sw, sh float64) {
	gx := c.px + c.pw - resizeGripSize
	gy := c.py + c.ph - resizeGripSize
	inGrip := mx >= gx && mx <= c.px+c.pw && my >= gy && my <= c.py+c.ph
	inTitle := mx >= c.px && mx <= c.px+c.pw && my >= c.py && my < c.py+titleBarH

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if inGrip {
			c.resizing = true
			c.resizeOX = mx
			c.resizeOY = my
		} else if inTitle {
			c.dragging = true
			c.dragOX = mx - c.px
			c.dragOY = my - c.py
		}
	}

	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		c.dragging = false
		c.resizing = false
	}

	if c.dragging {
		c.px = math.Max(0, math.Min(mx-c.dragOX, sw-c.pw))
		c.py = math.Max(0, math.Min(my-c.dragOY, sh-c.ph))
	}

	if c.resizing {
		c.pw += mx - c.resizeOX
		c.ph += my - c.resizeOY
		c.pw = math.Max(panelMinW, math.Min(c.pw, sw-c.px))
		c.ph = math.Max(panelMinH, math.Min(c.ph, sh-c.py))
		c.resizeOX = mx
		c.resizeOY = my
	}
}

func (c *Console) updateScroll(mx, my float64) {
	if !c.hitTest(mx, my) {
		return
	}
	_, sy := ebiten.Wheel()
	if sy == 0 {
		return
	}
	c.scrollOff -= int(sy) * scrollLines
	maxOff := len(c.output) - c.visLines()
	if maxOff < 0 {
		maxOff = 0
	}
	if c.scrollOff > maxOff {
		c.scrollOff = maxOff
	}
	if c.scrollOff < 0 {
		c.scrollOff = 0
	}
}

func (c *Console) ins(ch rune) {
	c.input = append(c.input, 0)
	copy(c.input[c.cursor+1:], c.input[c.cursor:])
	c.input[c.cursor] = ch
	c.cursor++
	c.blink = 0
}

func (c *Console) doBackspace() {
	if c.hasSel() {
		c.deleteSelection()
		return
	}
	if c.cursor > 0 {
		c.input = append(c.input[:c.cursor-1], c.input[c.cursor:]...)
		c.cursor--
		c.blink = 0
	}
}

func (c *Console) doDelete() {
	if c.hasSel() {
		c.deleteSelection()
		return
	}
	if c.cursor < len(c.input) {
		c.input = append(c.input[:c.cursor], c.input[c.cursor+1:]...)
		c.blink = 0
	}
}

func (c *Console) repeatableKey(key ebiten.Key, fn func()) {
	if inpututil.IsKeyJustPressed(key) {
		fn()
		return
	}
	if ebiten.IsKeyPressed(key) {
		dur := inpututil.KeyPressDuration(key)
		if dur > 18 && dur%3 == 0 {
			fn()
		}
	}
}

func (c *Console) hasSel() bool {
	return c.selActive && c.selAnchor >= 0 && c.selAnchor != c.selCursor
}

func (c *Console) selRange() (int, int) {
	s, e := c.selAnchor, c.selCursor
	if s > e {
		s, e = e, s
	}
	if s < 0 {
		s = 0
	}
	if e > len(c.input) {
		e = len(c.input)
	}
	return s, e
}

func (c *Console) selText() string {
	if !c.hasSel() {
		return ""
	}
	s, e := c.selRange()
	return string(c.input[s:e])
}

func (c *Console) deleteSelection() {
	if !c.hasSel() {
		return
	}
	s, e := c.selRange()
	c.input = append(c.input[:s], c.input[e:]...)
	c.cursor = s
	c.selClear()
	c.blink = 0
}

func (c *Console) selClear() {
	c.selAnchor = -1
	c.selCursor = -1
	c.selActive = false
}

func (c *Console) selExtend(dir int) {
	if !c.selActive {
		c.selAnchor = c.cursor
		c.selCursor = c.cursor
		c.selActive = true
	}
	c.cursor += dir
	if c.cursor < 0 {
		c.cursor = 0
	}
	if c.cursor > len(c.input) {
		c.cursor = len(c.input)
	}
	c.selCursor = c.cursor
}

func (c *Console) selExtendTo(pos int) {
	if !c.selActive {
		c.selAnchor = c.cursor
		c.selActive = true
	}
	c.selCursor = pos
}

var internalClipboard string

func (c *Console) doCopy() {
	if text := c.selText(); text != "" {
		internalClipboard = text
	}
}

func (c *Console) doPaste() {
	if internalClipboard == "" {
		return
	}
	c.deleteSelection()
	text := strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ", "\t", " ").Replace(internalClipboard)
	for _, ch := range text {
		c.ins(ch)
	}
}

func (c *Console) emit(text string, clr color.RGBA) {
	c.output = append(c.output, line{text: text, clr: clr})
	if len(c.output) > maxOutputLines {
		c.output = c.output[len(c.output)-maxOutputLines:]
	}
	c.scrollOff = 0
}

func (c *Console) submit() {
	raw := strings.TrimSpace(string(c.input))
	c.input = c.input[:0]
	c.cursor = 0
	c.selClear()
	c.suggestion = ""
	c.tabOptions = nil
	c.tabBase = ""
	if raw == "" {
		return
	}
	c.emit(inputPrefix+raw, colPrompt)
	if len(c.cmdHistory) == 0 || c.cmdHistory[len(c.cmdHistory)-1] != raw {
		c.cmdHistory = append(c.cmdHistory, raw)
		if len(c.cmdHistory) > maxCmdHistory {
			c.cmdHistory = c.cmdHistory[1:]
		}
	}
	c.cmdIdx = len(c.cmdHistory)
	c.execute(raw)
}

func (c *Console) execute(raw string) {
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return
	}
	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/help":
		c.emit("Commands:", colPrompt)
		c.emit("  /set KEY VALUE  — set a config variable", colText)
		c.emit("  /get KEY        — show current value", colText)
		c.emit("  /list [PREFIX]  — list config keys", colText)
		c.emit("  /reset          — reset config to defaults", colText)
		c.emit("  /clear          — clear console output", colText)
		c.emit("  /help           — show this help", colText)
		c.emit("", colText)
		c.emit("Shorthand: KEY VALUE  (same as /set KEY VALUE)", colHint)
		c.emit("           KEY        (same as /get KEY)", colHint)
		c.emit("", colText)
		c.emit("Resolution: 480p, 720p, 900p, 1080p, 1440p, 4k", colHint)
		c.emit("  e.g.  RESOLUTION 1080p", colHint)
		c.emit("", colText)
		c.emit("Tab autocomplete. Up/Down history. Shift+Arrow select.", colHint)
		c.emit("Ctrl+C copy, Ctrl+V paste, Ctrl+A select all.", colHint)
		c.emit("Drag title bar to move. Drag corner grip to resize.", colHint)

	case "/set":
		if len(parts) < 3 {
			c.emit("usage: /set KEY VALUE", colError)
			return
		}
		key := strings.TrimSpace(parts[1])
		value := strings.TrimSpace(strings.Join(parts[2:], " "))
		result := config.SetValue(c.cfg, key, value)
		c.emit(result, colText)
		c.notifySet(strings.ToUpper(key))

	case "/get":
		if len(parts) < 2 {
			c.emit("usage: /get KEY", colError)
			return
		}
		key := strings.TrimSpace(parts[1])
		val := config.GetValue(c.cfg, key)
		if val == "" {
			c.emit(fmt.Sprintf("unknown key: %s", strings.ToUpper(key)), colError)
		} else {
			c.emit(fmt.Sprintf("%s = %s", strings.ToUpper(key), val), colText)
		}

	case "/list":
		prefix := ""
		if len(parts) > 1 {
			prefix = strings.ToUpper(parts[1])
		}
		count := 0
		for _, k := range config.AllKeys() {
			if prefix == "" || strings.HasPrefix(k, prefix) {
				c.emit(fmt.Sprintf("  %s = %s", k, config.GetValue(c.cfg, k)), colText)
				count++
			}
		}
		if count == 0 {
			c.emit("no matching keys", colHint)
		}

	case "/reset":
		def := config.Default()
		*c.cfg = *def
		c.emit("config reset to defaults", colPrompt)

	case "/clear":
		c.output = c.output[:0]

	default:
		if strings.HasPrefix(cmd, "/") {
			c.emit(fmt.Sprintf("unknown command: %s (try /help)", cmd), colError)
			return
		}
		key := strings.ToUpper(parts[0])
		if config.GetValue(c.cfg, key) == "" {
			c.emit(fmt.Sprintf("unknown: %s (try /help)", raw), colError)
			return
		}
		if len(parts) >= 2 {
			value := strings.TrimSpace(strings.Join(parts[1:], " "))
			c.emit(config.SetValue(c.cfg, key, value), colText)
			c.notifySet(key)
		} else {
			c.emit(fmt.Sprintf("%s = %s", key, config.GetValue(c.cfg, key)), colText)
		}
	}
}

func (c *Console) historyUp() {
	if len(c.cmdHistory) == 0 {
		return
	}
	if c.cmdIdx > 0 {
		c.cmdIdx--
	}
	c.loadInput(c.cmdHistory[c.cmdIdx])
}

func (c *Console) historyDown() {
	if c.cmdIdx >= len(c.cmdHistory)-1 {
		c.cmdIdx = len(c.cmdHistory)
		c.input = c.input[:0]
		c.cursor = 0
		return
	}
	c.cmdIdx++
	c.loadInput(c.cmdHistory[c.cmdIdx])
}

func (c *Console) loadInput(s string) {
	c.input = []rune(s)
	c.cursor = len(c.input)
	c.selClear()
	c.blink = 0
}

func (c *Console) updateSuggestion() {
	c.suggestion = ""
	text := string(c.input)
	if text == "" {
		c.tabOptions = nil
		return
	}
	parts := strings.Fields(text)
	if len(parts) == 0 {
		c.tabOptions = nil
		return
	}

	trailingSpace := len(text) > 0 && text[len(text)-1] == ' '

	if len(parts) == 1 && !trailingSpace {
		tok := parts[0]
		if strings.HasPrefix(tok, "/") {
			lower := strings.ToLower(tok)
			for _, cmd := range []string{"/clear", "/get", "/help", "/list", "/reset", "/set"} {
				if strings.HasPrefix(cmd, lower) && cmd != lower {
					c.suggestion = cmd
					return
				}
			}
			return
		}
		upper := strings.ToUpper(tok)
		for _, k := range c.allKeys {
			if strings.HasPrefix(k, upper) && k != upper {
				c.suggestion = k
				return
			}
		}
		return
	}

	cmd := strings.ToLower(parts[0])
	isSetCmd := cmd == "/set" || (!strings.HasPrefix(cmd, "/") && config.GetValue(c.cfg, strings.ToUpper(parts[0])) != "")

	if (cmd == "/set" || cmd == "/get" || cmd == "/list") && len(parts) == 2 && !trailingSpace {
		prefix := strings.ToUpper(parts[1])
		for _, k := range c.allKeys {
			if strings.HasPrefix(k, prefix) && k != prefix {
				c.suggestion = cmd + " " + k
				return
			}
		}
		return
	}

	if isSetCmd {
		var key string
		var valuePrefix string
		var cmdPrefix string

		if cmd == "/set" {
			if len(parts) == 2 && trailingSpace {
				key = strings.ToUpper(parts[1])
				valuePrefix = ""
				cmdPrefix = "/set " + key + " "
			} else if len(parts) >= 3 {
				key = strings.ToUpper(parts[1])
				valuePrefix = strings.Join(parts[2:], " ")
				cmdPrefix = "/set " + key + " "
			}
		} else {
			key = strings.ToUpper(parts[0])
			if len(parts) == 1 && trailingSpace {
				valuePrefix = ""
				cmdPrefix = key + " "
			} else if len(parts) >= 2 {
				valuePrefix = strings.Join(parts[1:], " ")
				cmdPrefix = key + " "
			}
		}

		if key != "" {
			opts := config.ValueOptions(key)
			if len(opts) > 0 {
				lower := strings.ToLower(valuePrefix)
				var matches []string
				for _, o := range opts {
					if strings.HasPrefix(strings.ToLower(o), lower) {
						matches = append(matches, o)
					}
				}
				newBase := cmdPrefix + valuePrefix
				if c.tabBase != newBase {
					c.tabOptions = matches
					c.tabIndex = 0
					c.tabBase = newBase
				} else {
					c.tabOptions = matches
				}
				if len(matches) > 0 {
					idx := c.tabIndex % len(matches)
					c.suggestion = cmdPrefix + matches[idx]
				}
				return
			}
		}
	}

	c.tabOptions = nil
}

func (c *Console) completeTab() {
	if len(c.tabOptions) > 1 {
		idx := c.tabIndex % len(c.tabOptions)
		text := string(c.input)
		parts := strings.Fields(text)
		cmd := strings.ToLower(parts[0])

		var key string
		var cmdPrefix string
		if cmd == "/set" && len(parts) >= 2 {
			key = strings.ToUpper(parts[1])
			cmdPrefix = "/set " + key + " "
		} else if !strings.HasPrefix(cmd, "/") && len(parts) >= 1 {
			key = strings.ToUpper(parts[0])
			cmdPrefix = key + " "
		}

		if key != "" && len(c.tabOptions) > 0 {
			chosen := c.tabOptions[idx]
			c.input = []rune(cmdPrefix + chosen)
			c.cursor = len(c.input)
			c.selClear()
			c.blink = 0
			c.tabIndex = (c.tabIndex + 1) % len(c.tabOptions)
			c.tabBase = cmdPrefix
			return
		}
	}

	if c.suggestion == "" {
		return
	}
	c.input = []rune(c.suggestion)
	c.cursor = len(c.input)
	c.selClear()
	c.suggestion = ""
	c.tabOptions = nil
	c.blink = 0
}


