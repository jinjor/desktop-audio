package audio

import (
	"encoding/json"
	"log"
	"math"
	"strconv"
)

// ----- Filter Kind ----- //

//go:generate go run ../gen/main.go -- filter_kind.gen.go
/*
generate-enum filterKind

filterNone none
filterLowPassFIR lowpass-fir
filterHighPassFIR highpass-fir
filterLowPass lowpass
filterHighPass highpass
filterBandPass1 bandpass-1
filterBandPass2 bandpass-2
filterNotch notch
filterPeaking peaking
filterLowShelf lowshelf
filterHighShelf highshelf

EOF
*/

// ----- Filter Params ----- //

type filterParams struct {
	enabled bool
	kind    int
	freq    float64
	q       float64
	gain    float64
	N       int
}

type filterJSON struct {
	Enabled bool    `json:"enabled"`
	Kind    string  `json:"kind"`
	Freq    float64 `json:"freq"`
	Q       float64 `json:"q"`
	Gain    float64 `json:"gain"`
}

func (f *filterParams) applyJSON(data json.RawMessage) {
	var j filterJSON
	err := json.Unmarshal(data, &j)
	if err != nil {
		log.Println("failed to apply JSON to filter")
		return
	}
	f.enabled = j.Enabled
	f.kind = filterKindFromString(j.Kind)
	f.freq = j.Freq
	f.q = j.Q
	f.gain = j.Gain
}
func (f *filterParams) toJSON() json.RawMessage {
	return toRawMessage(&filterJSON{
		Enabled: f.enabled,
		Kind:    filterKindToString(f.kind),
		Freq:    f.freq,
		Q:       f.q,
		Gain:    f.gain,
	})
}
func (f *filterParams) set(key string, value string) error {
	switch key {
	case "enabled":
		f.enabled = value == "true"
	case "kind":
		f.kind = filterKindFromString(value)
	case "freq":
		freq, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		f.freq = freq
	case "q":
		q, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		f.q = q
	case "gain":
		gain, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		f.gain = gain
	}
	return nil
}

// ----- Filter ----- //

type filter struct {
	enabled bool
	kind    int
	freq    float64
	q       float64
	gain    float64
	N       int
	a       []float64 // feedforward
	b       []float64 // feedback
	past    []float64
}

func (f *filter) applyParams(p *filterParams) {
	f.enabled = p.enabled
	f.kind = p.kind
	f.freq = p.freq
	f.q = p.q
	f.gain = p.gain
	f.N = p.N
}
func (f *filter) step(in float64, freqRatio float64, envelopes []*envelope) float64 {
	if !f.enabled {
		return in
	}
	qExponent := 1.0
	gainRatio := 1.0
	for _, envelope := range envelopes {
		if !envelope.enabled {
			continue
		}
		if envelope.destination == destFilterFreq {
			freqRatio *= math.Pow(16.0, envelope.getValue())
		}
		if envelope.destination == destFilterQ || envelope.destination == destFilterQ0V {
			qExponent *= envelope.getValue()
		}
		if envelope.destination == destFilterGain || envelope.destination == destFilterGain0V {
			gainRatio *= envelope.getValue()
		}
	}
	f.a, f.b = makeH(f.a, f.b, f.kind, f.N, f.freq*freqRatio, math.Pow(f.q, qExponent), f.gain*gainRatio)
	pastLength := int(math.Max(float64(len(f.a)-1), float64(len(f.b))))
	if len(f.past) < pastLength {
		f.past = make([]float64, pastLength)
	}
	return calcFilterOneSample(in, f.a, f.b, f.past)
}
func makeH(
	feedforward []float64,
	feedback []float64,
	kind int,
	N int,
	freq float64,
	q float64,
	gain float64,
) ([]float64, []float64) {
	fc := freq / sampleRate
	switch kind {
	case filterLowPassFIR:
		return makeFIRLowpassH(feedforward, feedback, N, fc, hamming)
	case filterHighPassFIR:
		return makeFIRHighpassH(feedforward, feedback, N, fc, hamming)
	case filterLowPass:
		return makeBiquadLowpassH(feedforward, feedback, fc, q)
	case filterHighPass:
		return makeBiquadHighpassH(feedforward, feedback, fc, q)
	case filterBandPass1:
		return makeBiquadBandpass1H(feedforward, feedback, fc, q)
	case filterBandPass2:
		return makeBiquadBandpass2H(feedforward, feedback, fc, q)
	case filterNotch:
		return makeBiquadNotchH(feedforward, feedback, fc, q)
	case filterPeaking:
		return makeBiquadPeakingEQH(feedforward, feedback, fc, q, gain)
	case filterLowShelf:
		return makeBiquadLowShelfH(feedforward, feedback, fc, q, gain)
	case filterHighShelf:
		return makeBiquadHighShelfH(feedforward, feedback, fc, q, gain)
	case filterNone:
		fallthrough
	default:
		return makeNoFilterH(feedforward, feedback)
	}
}
func calcFilterOneSample(in float64, a []float64, b []float64, past []float64) float64 {
	// apply b
	for j := 0; j < len(b); j++ {
		in -= past[j] * b[j]
	}
	// apply a
	o := in * a[0]
	for j := 1; j < len(a); j++ {
		o += past[j-1] * a[j]
	}
	// unshift f.past
	for j := len(past) - 2; j >= 0; j-- {
		past[j+1] = past[j]
	}
	if len(past) > 0 {
		past[0] = in
	}
	return o
}

// ----- Calculation ----- //

func makeNewSliceIfLengthsAreNotTheSame(slice []float64, expectedLength int) []float64 {
	if len(slice) == expectedLength {
		return slice
	}
	if cap(slice) >= expectedLength {
		return slice[:expectedLength]
	}
	return make([]float64, expectedLength)
}

func makeNoFilterH(feedforward []float64, feedback []float64) ([]float64, []float64) {
	feedforward = makeNewSliceIfLengthsAreNotTheSame(feedforward, 1)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 0)
	feedforward[0] = 1
	return feedforward, feedback
}

func makeFIRLowpassH(feedforward []float64, feedback []float64, N int, fc float64, windowFunc func(float64) float64) ([]float64, []float64) {
	w0 := 2 * math.Pi * fc
	if N%2 != 0 {
		log.Panicf("N should be even")
	}
	h := makeNewSliceIfLengthsAreNotTheSame(feedforward, N+1)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 0)
	for i := 0; i <= N; i++ {
		n := float64(i - N/2)
		h[i] = 2 * fc * sinc(w0*n)
	}
	applyWindow(h, windowFunc)
	return h, feedback
}

func makeFIRHighpassH(feedforward []float64, feedback []float64, N int, fc float64, windowFunc func(float64) float64) ([]float64, []float64) {
	w0 := 2 * math.Pi * fc
	if N%2 != 0 {
		log.Panicf("N should be even")
	}
	h := makeNewSliceIfLengthsAreNotTheSame(feedforward, N+1)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 0)
	for i := 0; i <= N; i++ {
		n := float64(i - N/2)
		h[i] = sinc(math.Pi*n) - 2*fc*sinc(w0*n)
	}
	applyWindow(h, windowFunc)
	return h, feedback
}

func makeBiquadLowpassH(feedforward []float64, feedback []float64, fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := (1 - math.Cos(w0)) / 2
	b1 := (1 - math.Cos(w0))
	b2 := (1 - math.Cos(w0)) / 2
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	feedforward = makeNewSliceIfLengthsAreNotTheSame(feedforward, 3)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 2)
	feedforward[0], feedforward[1], feedforward[2] = b0/a0, b1/a0, b2/a0
	feedback[0], feedback[1] = a1/a0, a2/a0
	return feedforward, feedback
}
func makeBiquadHighpassH(feedforward []float64, feedback []float64, fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := (1 + math.Cos(w0)) / 2
	b1 := -(1 + math.Cos(w0))
	b2 := (1 + math.Cos(w0)) / 2
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	feedforward = makeNewSliceIfLengthsAreNotTheSame(feedforward, 3)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 2)
	feedforward[0], feedforward[1], feedforward[2] = b0/a0, b1/a0, b2/a0
	feedback[0], feedback[1] = a1/a0, a2/a0
	return feedforward, feedback
}

func makeBiquadBandpass1H(feedforward []float64, feedback []float64, fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := math.Sin(w0) / 2
	b1 := 0.0
	b2 := -math.Sin(w0) / 2
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	feedforward = makeNewSliceIfLengthsAreNotTheSame(feedforward, 3)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 2)
	feedforward[0], feedforward[1], feedforward[2] = b0/a0, b1/a0, b2/a0
	feedback[0], feedback[1] = a1/a0, a2/a0
	return feedforward, feedback
}

func makeBiquadBandpass2H(feedforward []float64, feedback []float64, fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := alpha
	b1 := 0.0
	b2 := -alpha
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	feedforward = makeNewSliceIfLengthsAreNotTheSame(feedforward, 3)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 2)
	feedforward[0], feedforward[1], feedforward[2] = b0/a0, b1/a0, b2/a0
	feedback[0], feedback[1] = a1/a0, a2/a0
	return feedforward, feedback
}

func makeBiquadNotchH(feedforward []float64, feedback []float64, fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := 1.0
	b1 := -2 * math.Cos(w0)
	b2 := 1.0
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	feedforward = makeNewSliceIfLengthsAreNotTheSame(feedforward, 3)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 2)
	feedforward[0], feedforward[1], feedforward[2] = b0/a0, b1/a0, b2/a0
	feedback[0], feedback[1] = a1/a0, a2/a0
	return feedforward, feedback
}

func makeBiquadAllpassH(feedforward []float64, feedback []float64, fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := 1 - alpha
	b1 := -2 * math.Cos(w0)
	b2 := 1 + alpha
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	feedforward = makeNewSliceIfLengthsAreNotTheSame(feedforward, 3)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 2)
	feedforward[0], feedforward[1], feedforward[2] = b0/a0, b1/a0, b2/a0
	feedback[0], feedback[1] = a1/a0, a2/a0
	return feedforward, feedback
}

func makeBiquadPeakingEQH(feedforward []float64, feedback []float64, fc float64, q float64, dBgain float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	A := math.Pow(10, dBgain/40)
	b0 := 1 + alpha*A
	b1 := -2 * math.Cos(w0)
	b2 := 1 - alpha*A
	a0 := 1 + alpha/A
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha/A
	feedforward = makeNewSliceIfLengthsAreNotTheSame(feedforward, 3)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 2)
	feedforward[0], feedforward[1], feedforward[2] = b0/a0, b1/a0, b2/a0
	feedback[0], feedback[1] = a1/a0, a2/a0
	return feedforward, feedback
}

func makeBiquadLowShelfH(feedforward []float64, feedback []float64, fc float64, q float64, dBgain float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	A := math.Pow(10, dBgain/40)
	b0 := A * ((A + 1) - (A-1)*math.Cos(w0) + 2*math.Sqrt(A)*alpha)
	b1 := 2 * A * ((A - 1) - (A+1)*math.Cos(w0))
	b2 := A * ((A + 1) - (A-1)*math.Cos(w0) - 2*math.Sqrt(A)*alpha)
	a0 := (A + 1) + (A-1)*math.Cos(w0) + 2*math.Sqrt(A)*alpha
	a1 := -2 * ((A - 1) + (A+1)*math.Cos(w0))
	a2 := (A + 1) + (A-1)*math.Cos(w0) - 2*math.Sqrt(A)*alpha
	feedforward = makeNewSliceIfLengthsAreNotTheSame(feedforward, 3)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 2)
	feedforward[0], feedforward[1], feedforward[2] = b0/a0, b1/a0, b2/a0
	feedback[0], feedback[1] = a1/a0, a2/a0
	return feedforward, feedback
}

func makeBiquadHighShelfH(feedforward []float64, feedback []float64, fc float64, q float64, dBgain float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	A := math.Pow(10, dBgain/40)
	b0 := A * ((A + 1) + (A-1)*math.Cos(w0) + 2*math.Sqrt(A)*alpha)
	b1 := -2 * A * ((A - 1) + (A+1)*math.Cos(w0))
	b2 := A * ((A + 1) + (A-1)*math.Cos(w0) - 2*math.Sqrt(A)*alpha)
	a0 := (A + 1) - (A-1)*math.Cos(w0) + 2*math.Sqrt(A)*alpha
	a1 := 2 * ((A - 1) - (A+1)*math.Cos(w0))
	a2 := (A + 1) - (A-1)*math.Cos(w0) - 2*math.Sqrt(A)*alpha
	feedforward = makeNewSliceIfLengthsAreNotTheSame(feedforward, 3)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 2)
	feedforward[0], feedforward[1], feedforward[2] = b0/a0, b1/a0, b2/a0
	feedback[0], feedback[1] = a1/a0, a2/a0
	return feedforward, feedback
}

func sinc(x float64) float64 {
	if math.Abs(x) < 0.000000001 {
		return 1
	}
	return math.Sin(x) / x
}
