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

// Generation

func calcFourierSeriesAtPhase(partials int, phase float64, calcFourierPartialAtPhase func(n int, phase float64) float64) float64 {
	value := 0.0
	for i := 1; i <= partials; i++ {
		value += calcFourierPartialAtPhase(i, phase)
	}
	return value
}
func makeBandLimitedTableForGivenNumberOfPartials(samples int, partials int, calcFourierPartialAtPhase func(n int, phase float64) float64) *wavetable {
	return newWaveTable(samples, func(phase float64) float64 {
		return calcFourierSeriesAtPhase(partials, phase, calcFourierPartialAtPhase)
	})
}
func makeBandLimitedTableWithMaxNumbersOfPartialsAtNote(samples int, note int, calcFourierPartialAtPhase func(n int, phase float64) float64) *wavetable {
	freq := 442 * math.Pow(2, float64(note-69)/12)
	partials := int(sampleRate / 2 / freq)
	return makeBandLimitedTableForGivenNumberOfPartials(samples, partials, calcFourierPartialAtPhase)
}
func makeBandLimitedTablesForAllNotes(samples int, calcFourierPartialAtPhase func(n int, phase float64) float64) []*wavetable {
	tables := make([]*wavetable, 128)
	for i := 0; i < 128; i++ {
		tables[i] = makeBandLimitedTableWithMaxNumbersOfPartialsAtNote(samples, i, calcFourierPartialAtPhase)
	}
	return tables
}
func calcPartialSquareAtPhase(n int, phase float64) float64 {
	if n%2 == 1 {
		x := float64(n)
		return math.Sin(x*phase) / x
	}
	return 0.0
}
func calcPartialSawAtPhase(n int, phase float64) float64 {
	x := float64(n)
	return math.Sin(x*phase) / x
}

// TODO: Store somewhere and load
// var squarebltable = makeBandLimitedTablesForAllNotes(4096, calcPartialSquareAtPhase)
// var sawbltable = makeBandLimitedTablesForAllNotes(4096, calcPartialSawAtPhase)
