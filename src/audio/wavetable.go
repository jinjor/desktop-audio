package audio

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

type wavetable struct {
	values []float64
}

func newWavetable(cap int) *wavetable {
	return &wavetable{
		values: make([]float64, 0, cap),
	}
}
func (wt *wavetable) generate(samples int, phaseToValue func(phase float64) float64) error {
	if samples > cap(wt.values) {
		return fmt.Errorf("capacity exceeded")
	}
	wt.values = wt.values[0:samples]
	for i := 0; i < samples; i++ {
		phase := 2.0 * math.Pi / float64(samples) * float64(i)
		wt.values[i] = phaseToValue(phase)
	}
	return nil
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
func (wt *wavetable) makeBandLimitedTableForGivenNumberOfPartials(samples int, partials int, calcFourierPartialAtPhase func(n int, phase float64) float64) {
	wt.generate(samples, func(phase float64) float64 {
		value := 0.0
		for i := 1; i <= partials; i++ {
			value += calcFourierPartialAtPhase(i, phase)
		}
		return value
	})
}
func (wt *wavetable) makeBandLimitedTableWithMaxNumbersOfPartialsAtNote(samples int, note int, calcFourierPartialAtPhase func(n int, phase float64) float64) {
	freq := baseFreq * math.Pow(2, float64(note-69)/12)
	partials := int(sampleRate / 2 / freq)
	wt.makeBandLimitedTableForGivenNumberOfPartials(samples, partials, calcFourierPartialAtPhase)
}

// WavetableSet ...
type WavetableSet struct {
	tables []*wavetable
}

// NewWavetableSet ...
func NewWavetableSet(tableCap int, sampleCap int) *WavetableSet {
	tables := make([]*wavetable, tableCap)
	for i := 0; i < tableCap; i++ {
		tables[i] = newWavetable(sampleCap)
	}
	return &WavetableSet{
		tables: tables,
	}
}

// MakeBandLimitedTablesForAllNotes ...
func (wts *WavetableSet) MakeBandLimitedTablesForAllNotes(samples int, calcFourierPartialAtPhase func(n int, phase float64) float64) error {
	if cap(wts.tables) < 128 {
		return fmt.Errorf("capacity of tables exceeded")
	}
	wts.tables = wts.tables[0:128]
	for i := 0; i < 128; i++ {
		wts.tables[i].makeBandLimitedTableWithMaxNumbersOfPartialsAtNote(samples, i, calcFourierPartialAtPhase)
	}
	return nil
}

// IO
//   all = { number_of_tables int32, tables []table }
//   table = { number_of_samples int32, samples []float64 }

// Save ...
func (wts *WavetableSet) Save(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		return err
	}
	numTables := int32(len(wts.tables))
	err = binary.Write(file, binary.BigEndian, numTables)
	if err != nil {
		return err
	}
	for _, wt := range wts.tables {
		numSamples := int32(len(wt.values))
		err = binary.Write(file, binary.BigEndian, numSamples)
		if err != nil {
			return err
		}
		for _, value := range wt.values {
			err := binary.Write(file, binary.BigEndian, value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Load ...
func (wts *WavetableSet) Load(path string) error {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return err
	}
	var numTables int32
	err = binary.Read(file, binary.BigEndian, &numTables)
	if err != nil {
		return err
	}
	if int(numTables) > cap(wts.tables) {
		return fmt.Errorf("number of tables exceeded")
	}
	wts.tables = wts.tables[0:numTables]
	for _, wt := range wts.tables {
		var numSamples int32
		err = binary.Read(file, binary.BigEndian, &numSamples)
		if err != nil {
			return err
		}
		if int(numSamples) > cap(wt.values) {
			return fmt.Errorf("number of samples exceeded")
		}
		wt.values = wt.values[0:numSamples]
		for i := range wt.values {
			err = binary.Read(file, binary.BigEndian, &wt.values[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}
