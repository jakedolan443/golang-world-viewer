
package timecontrol

type TimeControl struct {
	Speed    float64
	GameTime float64
	Tick     uint64
	Paused   bool

	speedOptions []float64
	speedIndex   int
}

func New(speedOptions []float64, defaultSpeed float64) *TimeControl {
	idx := 0
	for i, s := range speedOptions {
		if s == defaultSpeed {
			idx = i
			break
		}
	}
	return &TimeControl{
		Speed:        speedOptions[idx],
		speedOptions: speedOptions,
		speedIndex:   idx,
	}
}

func (tc *TimeControl) Update(dt float64) {
	if tc.Paused {
		return
	}
	tc.GameTime += dt * tc.Speed
	tc.Tick++
}

func (tc *TimeControl) SetSpeed(speed float64) {
	tc.Speed = speed
	for i, s := range tc.speedOptions {
		if s == speed {
			tc.speedIndex = i
			return
		}
	}
}

func (tc *TimeControl) CycleSpeed() {
	tc.speedIndex = (tc.speedIndex + 1) % len(tc.speedOptions)
	tc.Speed = tc.speedOptions[tc.speedIndex]
}

func (tc *TimeControl) TogglePause() {
	tc.Paused = !tc.Paused
}
