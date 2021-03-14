package audio

// ----- Modulation -----

type modulation struct {
	oscVolumeRatio      []float64
	freqRatio           float64
	phaseShift          float64
	ampRatio            float64
	noteFilterFreqRatio float64
	noteFilterQExponent float64
	noteFilterGainRatio float64
	filterFreqRatio     float64
	filterQExponent     float64
	filterGainRatio     float64
	lfoAmountGain       []float64
	lfoFreqRatio        []float64
}

func newModulation() *modulation {
	m := &modulation{
		oscVolumeRatio: make([]float64, 2),
		lfoAmountGain:  make([]float64, 3),
		lfoFreqRatio:   make([]float64, 3),
	}
	m.init()
	return m
}
func (m *modulation) init() {
	m.oscVolumeRatio[0] = 1.0
	m.oscVolumeRatio[1] = 1.0
	m.freqRatio = 1.0
	m.phaseShift = 0.0
	m.ampRatio = 1.0
	m.noteFilterFreqRatio = 1.0
	m.noteFilterQExponent = 1.0
	m.noteFilterGainRatio = 1.0
	m.filterFreqRatio = 1.0
	m.filterQExponent = 1.0
	m.filterGainRatio = 1.0
	m.lfoAmountGain[0] = 1.0
	m.lfoAmountGain[1] = 1.0
	m.lfoAmountGain[2] = 1.0
	m.lfoFreqRatio[0] = 1.0
	m.lfoFreqRatio[1] = 1.0
	m.lfoFreqRatio[2] = 1.0
}
