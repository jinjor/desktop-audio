package audio

import "math"

// ----- Transition Kind ----- //

const (
	transitionNone = iota
	transitionLinear
	transitionExponential
)

// ----- Transitive Value ----- //

type transitiveValue struct {
	kind         int
	duration     float64 // ms
	endThreshold float64
	initialValue float64
	targetValue  float64
	value        float64
	pos          int
}

func (tv *transitiveValue) lenear(duration float64, targetValue float64) {
	tv.kind = transitionLinear
	tv.duration = duration
	tv.endThreshold = 0
	tv.pos = 0
	tv.initialValue = tv.value
	tv.targetValue = targetValue
}
func (tv *transitiveValue) exponential(duration float64, targetValue float64, endThreshold float64) {
	tv.kind = transitionExponential
	tv.duration = duration
	tv.endThreshold = endThreshold
	tv.pos = 0
	tv.initialValue = tv.value
	tv.targetValue = targetValue
}
func (tv *transitiveValue) step() bool {
	ended := false
	switch tv.kind {
	case transitionLinear:
		phaseTime := float64(tv.pos) * secPerSample * 1000 // ms
		if phaseTime >= float64(tv.duration) {
			tv.end()
			ended = true
		} else {
			t := phaseTime / float64(tv.duration)
			tv.value = t*tv.targetValue + (1-t)*tv.initialValue // TODO: don't use the same attack time
			tv.pos++
		}
	case transitionExponential:
		phaseTime := float64(tv.pos) * secPerSample * 1000 // ms
		t := phaseTime / float64(tv.duration)
		tv.value = setTargetAtTime(tv.initialValue, tv.targetValue, t)
		if math.Abs(tv.value-tv.targetValue) < tv.endThreshold {
			tv.end()
			ended = true
		} else {
			tv.pos++
		}
	case transitionNone:

	}
	return ended
}
func (tv *transitiveValue) end() {
	tv.kind = transitionNone
	tv.value = tv.targetValue
	tv.pos = 0
}

// 63% closer to target when pos=1.0
func setTargetAtTime(initialValue float64, targetValue float64, pos float64) float64 {
	return targetValue + (initialValue-targetValue)*math.Exp(-pos)
}
