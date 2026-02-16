
package layer

import (
	"math"
	"runtime"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"

	"shipping/internal/mapview"
	"shipping/internal/timecontrol"
)

const (
	noiseTableSize     = 512
	noiseTableMask     = noiseTableSize - 1
	noiseTableScale    = 20.0
	noiseTableInvScale = float64(noiseTableSize) / noiseTableScale
)

var (
	noiseTableMain   [noiseTableSize * noiseTableSize]float64
	noiseTableWarp   [noiseTableSize * noiseTableSize]float64
	noiseTableDetail [noiseTableSize * noiseTableSize]float64
	perm             [512]int
	grad2            = [12][2]float64{
		{1, 1}, {-1, 1}, {1, -1}, {-1, -1},
		{1, 0}, {-1, 0}, {0, 1}, {0, -1},
		{1, 1}, {-1, 1}, {1, -1}, {-1, -1},
	}
)

func init() {
	p := [256]int{
		151, 160, 137, 91, 90, 15, 131, 13, 201, 95, 96, 53, 194, 233, 7, 225,
		140, 36, 103, 30, 69, 142, 8, 99, 37, 240, 21, 10, 23, 190, 6, 148,
		247, 120, 234, 75, 0, 26, 197, 62, 94, 252, 219, 203, 117, 35, 11, 32,
		57, 177, 33, 88, 237, 149, 56, 87, 174, 20, 125, 136, 171, 168, 68, 175,
		74, 165, 71, 134, 139, 48, 27, 166, 77, 146, 158, 231, 83, 111, 229, 122,
		60, 211, 133, 230, 220, 105, 92, 41, 55, 46, 245, 40, 244, 102, 143, 54,
		65, 25, 63, 161, 1, 216, 80, 73, 209, 76, 132, 187, 208, 89, 18, 169,
		200, 196, 135, 130, 116, 188, 159, 86, 164, 100, 109, 198, 173, 186, 3, 64,
		52, 217, 226, 250, 124, 123, 5, 202, 38, 147, 118, 126, 255, 82, 85, 212,
		207, 206, 59, 227, 47, 16, 58, 17, 182, 189, 28, 42, 223, 183, 170, 213,
		119, 248, 152, 2, 44, 154, 163, 70, 221, 153, 101, 155, 167, 43, 172, 9,
		129, 22, 39, 253, 19, 98, 108, 110, 79, 113, 224, 232, 178, 185, 112, 104,
		218, 246, 97, 228, 251, 34, 242, 193, 238, 210, 144, 12, 191, 179, 162, 241,
		81, 51, 145, 235, 249, 14, 239, 107, 49, 192, 214, 31, 181, 199, 106, 157,
		184, 84, 204, 176, 115, 121, 50, 45, 127, 4, 150, 254, 138, 236, 205, 93,
		222, 114, 67, 29, 24, 72, 243, 141, 128, 195, 78, 66, 215, 61, 156, 180,
	}
	for i := 0; i < 512; i++ {
		perm[i] = p[i&255]
	}
	step := noiseTableScale / float64(noiseTableSize)
	for row := 0; row < noiseTableSize; row++ {
		y := float64(row) * step
		off := row * noiseTableSize
		for col := 0; col < noiseTableSize; col++ {
			x := float64(col) * step
			noiseTableMain[off+col] = fbm(x, y, 6)
			noiseTableWarp[off+col] = fbm(x, y, 3)
			noiseTableDetail[off+col] = fbm(x, y, 7)
		}
	}
}

//go:nosplit
func sampleNoise(table *[noiseTableSize * noiseTableSize]float64, x, y float64) float64 {
	tx := x*noiseTableInvScale + float64(noiseTableSize*256)
	ty := y*noiseTableInvScale + float64(noiseTableSize*256)
	ix, iy := int(tx), int(ty)
	fx, fy := tx-float64(ix), ty-float64(iy)
	x0, y0 := ix&noiseTableMask, iy&noiseTableMask
	x1, y1 := (x0+1)&noiseTableMask, (y0+1)&noiseTableMask
	r0, r1 := y0*noiseTableSize, y1*noiseTableSize
	top := table[r0+x0] + (table[r0+x1]-table[r0+x0])*fx
	bot := table[r1+x0] + (table[r1+x1]-table[r1+x0])*fx
	return top + (bot-top)*fy
}

func sampleNoiseWarp(x, y float64) float64   { return sampleNoise(&noiseTableWarp, x, y) }
func sampleNoiseMain(x, y float64) float64   { return sampleNoise(&noiseTableMain, x, y) }
func sampleNoiseDetail(x, y float64) float64 { return sampleNoise(&noiseTableDetail, x, y) }

func latCoverage(lat float64) float64 {
	a := math.Abs(lat)
	itcz := math.Exp(-a * a / 98)
	sub := 0.8 * math.Exp(-(a-22)*(a-22)/72)
	mid := 0.75 * math.Exp(-(a-50)*(a-50)/200)
	polar := 0.4 * smoothstep(65, 80, a)
	c := math.Max(itcz-sub, 0) + mid + polar
	if c > 1 {
		c = 1
	}
	return 0.05 + 0.95*c
}

//go:nosplit
func fastPow07(x float64) float64 {
	if x <= 0 {
		return 0
	}
	if x >= 1 {
		return 1
	}
	return math.Sqrt(x) * (1 + 4*x) / (4 + x)
}

func smoothstep(e0, e1, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-e0)/(e1-e0)))
	return t * t * (3 - 2*t)
}

func clampByte(v float64) byte {
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 255
	}
	return byte(v*255 + 0.5)
}

func fbm(x, y float64, octaves int) float64 {
	val, amp, freq, maxA := 0.0, 1.0, 1.0, 0.0
	for i := 0; i < octaves; i++ {
		val += simplex2D(x*freq, y*freq) * amp
		maxA += amp
		amp *= 0.5
		freq *= 2.0
	}
	return val / maxA
}

func simplex2D(xin, yin float64) float64 {
	const (
		f2 = 0.3660254037844386
		g2 = 0.21132486540518713
	)
	s := (xin + yin) * f2
	i := math.Floor(xin + s)
	j := math.Floor(yin + s)
	t := (i + j) * g2
	x0 := xin - (i - t)
	y0 := yin - (j - t)

	var i1, j1 int
	if x0 > y0 {
		i1, j1 = 1, 0
	} else {
		i1, j1 = 0, 1
	}

	x1 := x0 - float64(i1) + g2
	y1 := y0 - float64(j1) + g2
	x2 := x0 - 1.0 + 2.0*g2
	y2 := y0 - 1.0 + 2.0*g2

	ii, jj := int(i)&255, int(j)&255
	gi0 := perm[ii+perm[jj]] % 12
	gi1 := perm[ii+i1+perm[jj+j1]] % 12
	gi2 := perm[ii+1+perm[jj+1]] % 12

	n0, n1, n2 := 0.0, 0.0, 0.0
	if t0 := 0.5 - x0*x0 - y0*y0; t0 >= 0 {
		t0 *= t0
		n0 = t0 * t0 * (grad2[gi0][0]*x0 + grad2[gi0][1]*y0)
	}
	if t1 := 0.5 - x1*x1 - y1*y1; t1 >= 0 {
		t1 *= t1
		n1 = t1 * t1 * (grad2[gi1][0]*x1 + grad2[gi1][1]*y1)
	}
	if t2 := 0.5 - x2*x2 - y2*y2; t2 >= 0 {
		t2 *= t2
		n2 = t2 * t2 * (grad2[gi2][0]*x2 + grad2[gi2][1]*y2)
	}
	return 70.0 * (n0 + n1 + n2)
}

type colPrecomp struct {
	frac, wxCos, wxSinY     float64
	shadowFrac, shadowAngle float64
	sCos, sSinY             float64
}

type rowPrecomp struct {
	wy, northBlend, coverage, wyFreq float64
	doNorth, doSouth                 bool
	nWarpInY, nWarpInY2              float64
	sWarpInY, sWarpInY2              float64
}

type frameParams struct {
	nwx03, nwy03, swx03, swy03 float64
	northWindX, northWindY      float64
	southWindX, southWindY      float64
	detailWindX, detailWindY    float64
	warpAmt, warpAmt2           float64
	baseFreq, zoomAlpha         float64
	noiseRadius, sinYScale      float64
	shadowOffWY                 float64
	cols                        int
}

type OceanCloudLayer struct {
	zoomThreshold float64
	fadeDuration  float64
	cellWorld     float64
	targetPxPS    float64
	maxDim        int
	workerCount   int

	time      float64
	fadeAlpha float64
	offscreen *ebiten.Image
	offW      int
	offH      int
	pix       []byte
	colData   []colPrecomp
	rowData   []rowPrecomp
}

func NewOceanCloudLayer(zoomThreshold, fadeDuration, cellWorld, targetPxPS float64, maxDim, workerCount int) *OceanCloudLayer {
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
	}
	return &OceanCloudLayer{
		zoomThreshold: zoomThreshold,
		fadeDuration:  fadeDuration,
		cellWorld:     cellWorld,
		targetPxPS:    targetPxPS,
		maxDim:        maxDim,
		workerCount:   workerCount,
		fadeAlpha:     1.0,
	}
}

func (o *OceanCloudLayer) BelowLand() bool { return false }

func (o *OceanCloudLayer) Update(mv *mapview.MapView, tc *timecontrol.TimeControl) {
	o.time = tc.GameTime * 0.05
	fadeStep := 1.0 / (60.0 * o.fadeDuration)
	if mv.Zoom() > o.zoomThreshold {
		o.fadeAlpha = math.Max(0, o.fadeAlpha-fadeStep)
	} else {
		o.fadeAlpha = math.Min(1, o.fadeAlpha+fadeStep)
	}
}

func (o *OceanCloudLayer) Draw(mv *mapview.MapView, screen *ebiten.Image) {
	if o.fadeAlpha <= 0 {
		return
	}

	sw, sh := mv.ScreenSize()
	zoom := mv.Zoom()
	cellPx := o.cellWorld * zoom

	cellStep := 1
	ecw := o.cellWorld
	for cellPx*float64(cellStep) < 6 {
		cellStep++
		ecw = o.cellWorld * float64(cellStep)
	}
	ecp := ecw * zoom

	sub := int(math.Ceil(ecp / o.targetPxPS))
	if sub < 1 {
		sub = 1
	}
	sw2 := ecw / float64(sub)
	sp := sw2 * zoom

	camX, camY := mv.CameraPos()
	viewMinX := camX - sw/(2*zoom)
	viewMaxY := camY + sh/(2*zoom)

	gox := math.Floor(viewMinX/sw2) * sw2
	goy := math.Ceil(viewMaxY/sw2) * sw2
	sox := (gox - viewMinX) * zoom
	soy := (viewMaxY - goy) * zoom

	cols := int(math.Ceil(sw/sp)) + 2
	rows := int(math.Ceil(sh/sp)) + 2
	if cols > o.maxDim {
		cols = o.maxDim
	}
	if rows > o.maxDim {
		rows = o.maxDim
	}

	if o.offscreen == nil || o.offW != cols || o.offH != rows {
		o.offscreen = ebiten.NewImage(cols, rows)
		o.offW, o.offH = cols, rows
	}
	o.offscreen.Clear()

	total := cols * rows * 4
	if cap(o.pix) < total {
		o.pix = make([]byte, total)
	} else {
		o.pix = o.pix[:total]
	}
	for i := range o.pix {
		o.pix[i] = 0
	}

	bf := 0.022
	warpAmt := 14.0 * bf
	warpAmt2 := 8.0 * bf

	nwx := o.time * 0.06
	nwy := o.time * 0.0167
	swx := -o.time * 0.06
	swy := -o.time * 0.0167
	nwx03, nwy03 := nwx*0.3, nwy*0.3
	swx03, swy03 := swx*0.3, swy*0.3
	dwx := o.time * 0.025
	dwy := -o.time * 0.012

	wMinX, _, wMaxX, _ := mv.WorldBounds()
	ww := wMaxX - wMinX
	nr := ww * bf / (2 * math.Pi)
	sys := nr * 0.15
	invWW := 1.0 / ww
	sof := 0.12 * invWW

	if cap(o.colData) < cols {
		o.colData = make([]colPrecomp, cols)
	} else {
		o.colData = o.colData[:cols]
	}
	for col := 0; col < cols; col++ {
		wx := gox + (float64(col)+0.5)*sw2
		f := (wx - wMinX) * invWW
		f -= math.Floor(f)
		a := f * 2 * math.Pi
		sf := f + sof
		sf -= math.Floor(sf)
		sa := sf * 2 * math.Pi
		o.colData[col] = colPrecomp{
			frac: f, wxCos: math.Cos(a) * nr, wxSinY: math.Sin(a) * sys,
			shadowFrac: sf, shadowAngle: sa,
			sCos: math.Cos(sa) * nr, sSinY: math.Sin(sa) * sys,
		}
	}

	const bh = 10.0
	if cap(o.rowData) < rows {
		o.rowData = make([]rowPrecomp, rows)
	} else {
		o.rowData = o.rowData[:rows]
	}
	for row := 0; row < rows; row++ {
		wy := goy - (float64(row)+0.5)*sw2
		nb := smoothstep(-bh, bh, wy)
		wyf := wy * bf
		o.rowData[row] = rowPrecomp{
			wy: wy, northBlend: nb, doNorth: nb > 0.001, doSouth: nb < 0.999,
			coverage: latCoverage(wy), wyFreq: wyf,
			nWarpInY: wyf + nwy03 + 317.3, nWarpInY2: wyf + nwy03 + 541.7,
			sWarpInY: wyf + swy03 + 1317.3, sWarpInY2: wyf + swy03 + 1541.7,
		}
	}

	fp := frameParams{
		nwx03: nwx03, nwy03: nwy03, swx03: swx03, swy03: swy03,
		northWindX: nwx, northWindY: nwy, southWindX: swx, southWindY: swy,
		detailWindX: dwx, detailWindY: dwy,
		warpAmt: warpAmt, warpAmt2: warpAmt2,
		baseFreq: bf, zoomAlpha: o.fadeAlpha,
		noiseRadius: nr, sinYScale: sys,
		shadowOffWY: -0.20, cols: cols,
	}

	pix := o.pix
	cd := o.colData
	rd := o.rowData

	rowCh := make(chan int, rows)
	for i := 0; i < rows; i++ {
		rowCh <- i
	}
	close(rowCh)

	nw := o.workerCount
	if nw > rows {
		nw = rows
	}
	if nw < 1 {
		nw = 1
	}

	var wg sync.WaitGroup
	wg.Add(nw)
	for w := 0; w < nw; w++ {
		go func() {
			defer wg.Done()
			for row := range rowCh {
				renderCloudRow(row, pix, cd, &rd[row], &fp)
			}
		}()
	}
	wg.Wait()

	o.offscreen.WritePixels(pix)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(sp, sp)
	op.GeoM.Translate(sox, soy)
	op.Filter = ebiten.FilterLinear
	op.Blend = ebiten.BlendSourceOver
	screen.DrawImage(o.offscreen, op)
}

func renderCloudRow(row int, pix []byte, cd []colPrecomp, r *rowPrecomp, fp *frameParams) {
	cols := fp.cols
	rowOff := row * cols * 4
	nb := r.northBlend
	dn, ds := r.doNorth, r.doSouth
	cov := r.coverage
	wyf := r.wyFreq
	wy := r.wy
	wa, wa2 := fp.warpAmt, fp.warpAmt2
	za := fp.zoomAlpha

	const (
		noX, noY = 0.0, 0.0
		soX, soY = 1000.0, 1000.0
	)

	for col := 0; col < cols; col++ {
		c := &cd[col]
		wxC, wxS := c.wxCos, c.wxSinY

		var dN float64
		if dn {
			wx03, wy03 := fp.nwx03, fp.nwy03
			wpX := sampleNoiseWarp(wxC+wx03+113.7+noX, r.nWarpInY+wxS) * wa
			wpY := sampleNoiseWarp(wxC+wx03+743.2+noX, r.nWarpInY2+wxS) * wa * 0.85
			wpX += sampleNoiseDetail(wxC+wx03+227.1+noX+fp.detailWindX, wyf+wy03+419.6+noY+fp.detailWindY+wxS) * wa2
			wpY += sampleNoiseDetail(wxC+wx03+851.3+noX+fp.detailWindX, wyf+wy03+637.9+noY+fp.detailWindY+wxS) * wa2 * 1.15
			d := (sampleNoiseMain(wxC+wpX+fp.northWindX+noX, wyf+wpY+fp.northWindY+noY+wxS) + 1.0) * 0.5
			th := 0.62 - cov*0.25
			t := (d - th) / 0.38
			if t > 0 {
				if t > 1 {
					t = 1
				}
				d = t * t * (3 - 2*t)
				det := (sampleNoiseDetail(wxC+wpX*0.5+fp.northWindX*1.3+noX+937.4, wyf+wpY*0.5+fp.northWindY*1.3+noY+861.2+wxS) + 1.0) * 0.5
				dN = d * (0.50 + 0.50*det)
			}
		}

		var dS float64
		if ds {
			wx03, wy03 := fp.swx03, fp.swy03
			wpX := sampleNoiseWarp(wxC+wx03+113.7+soX, r.sWarpInY+wxS) * wa
			wpY := sampleNoiseWarp(wxC+wx03+743.2+soX, r.sWarpInY2+wxS) * wa * 0.85
			wpX += sampleNoiseDetail(wxC+wx03+227.1+soX+fp.detailWindX, wyf+wy03+419.6+soY+fp.detailWindY+wxS) * wa2
			wpY += sampleNoiseDetail(wxC+wx03+851.3+soX+fp.detailWindX, wyf+wy03+637.9+soY+fp.detailWindY+wxS) * wa2 * 1.15
			d := (sampleNoiseMain(wxC+wpX+fp.southWindX+soX, wyf+wpY+fp.southWindY+soY+wxS) + 1.0) * 0.5
			th := 0.62 - cov*0.25
			t := (d - th) / 0.38
			if t > 0 {
				if t > 1 {
					t = 1
				}
				d = t * t * (3 - 2*t)
				det := (sampleNoiseDetail(wxC+wpX*0.5+fp.southWindX*1.3+soX+937.4, wyf+wpY*0.5+fp.southWindY*1.3+soY+861.2+wxS) + 1.0) * 0.5
				dS = d * (0.50 + 0.50*det)
			}
		}

		density := dN*nb + dS*(1.0-nb)
		if density < 0.005 {
			continue
		}

		sC, sS := c.sCos, c.sSinY
		sWY := (wy + fp.shadowOffWY) * fp.baseFreq
		shWX := sampleNoiseWarp(sC+fp.nwx03+113.7, sWY+fp.nwy03+317.3+sS) * wa
		shWY := sampleNoiseWarp(sC+fp.nwx03+743.2, sWY+fp.nwy03+541.7+sS) * wa
		shD := (sampleNoiseMain(sC+shWX+fp.northWindX, sWY+shWY+fp.northWindY+sS) + 1.0) * 0.5
		if shD < 0.35 {
			shD = 0
		} else {
			shD = (shD - 0.35) / 0.40
			if shD > 1 {
				shD = 1
			}
			shD = shD * shD * (3 - 2*shD)
		}
		shA := shD * 0.04 * za

		shaped := fastPow07(density)
		mA := shaped * 0.16 * za
		if mA < 0.003 {
			continue
		}

		w := density * density * 0.05
		cR, cG, cB := 0.93+w, 0.95+w*0.5, 0.98
		t := shaped
		cr := 0.72 + (cR-0.72)*t
		cg := 0.80 + (cG-0.80)*t
		cb := 0.93 + (cB-0.93)*t

		om := 1 - mA
		outA := mA + shA*om
		outR := cr*mA + 0.06*shA*om
		outG := cg*mA + 0.09*shA*om
		outB := cb*mA + 0.18*shA*om

		idx := rowOff + col*4
		pix[idx] = clampByte(outR)
		pix[idx+1] = clampByte(outG)
		pix[idx+2] = clampByte(outB)
		pix[idx+3] = clampByte(outA)
	}
}
