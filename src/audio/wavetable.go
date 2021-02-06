package audio

import (
	"math"
)

type wavetable struct {
	values []float64
}

func newWaveTable(samples int, phaseToValue func(phase float64) float64) *wavetable {
	values := make([]float64, samples)
	for i := 0; i < samples; i++ {
		phase := 2.0 * math.Pi / float64(samples) * float64(i)
		values[i] = phaseToValue(phase)
	}
	return &wavetable{
		values: values,
	}
}

func (wt *wavetable) getAtPhase(phase float64) float64 {
	phase = math.Mod(phase, 2.0*math.Pi)
	length := len(wt.values)
	phasePerSample := 2.0 * math.Pi / float64(length)
	index := int(phase / phasePerSample)
	nextIndex := index + 1
	if nextIndex >= length {
		nextIndex = 0
	}
	mod := math.Mod(phase, phasePerSample)
	return wt.values[index]*(1-mod) + wt.values[nextIndex]*mod
}
