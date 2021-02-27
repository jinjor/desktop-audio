//go:generate go run ../gen/main.go -- enums.gen.go

package audio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/hajimehoshi/oto"
)

const (
	sampleRate      = 48000
	channelNum      = 2
	bitDepthInBytes = 2
	samplesPerCycle = 1024
	fftSize         = 2048 // multiple of samplesPerCycle
	maxPoly         = 128
)
const bytesPerSample = bitDepthInBytes * channelNum
const bufferSizeInBytes = samplesPerCycle * bytesPerSample // should be >= 4096
const secPerSample = 1.0 / sampleRate
const responseDelay = secPerSample * samplesPerCycle
const baseFreq = 442.0
const oscGain = 0.07

var fft = NewFFT(fftSize, false)

// ----- Utility ----- //

func now() float64 {
	return float64(time.Now().UnixNano()) / 1000 / 1000 / 1000
}
func positiveMod(a float64, b float64) float64 {
	if b < 0 {
		panic("b should not be negative")
	}
	for a < 0 {
		a += b
	}
	return math.Mod(a, b)
}
func noteToFreq(note int) float64 {
	return baseFreq * math.Pow(2, float64(note-69)/12)
}

// func freqToNote(freq float64) int {
// 	note := int(math.Log2(freq/baseFreq)*12.0) + 69
// 	if note < 0 {
// 		note = 0
// 	}
// 	if note >= 128 {
// 		note = 127
// 	}
// 	return note
// }

var freqs = makeFreqs()

func makeFreqs() []float64 {
	freqs := make([]float64, 128)
	for i := 0; i < 128; i++ {
		freqs[i] = noteToFreq(i)
	}
	return freqs
}
func freqToNote(freq float64) int {
	low := 0
	high := 128
	if freq < freqs[low] {
		return 0
	}
	if freq >= freqs[high-1] {
		return high - 1
	}
	for i := 0; i < 128; i++ {
		curr := (low + high) / 2
		if freq < freqs[curr] {
			high = curr
		} else {
			low = curr
		}
		if high-low <= 1 {
			return low
		}
	}
	panic("infinite loop in freqToNote()")
}
func toRawMessage(v interface{}) json.RawMessage {
	bytes, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return json.RawMessage(bytes)
}

// ----- MIDI Event ----- //

type midiEvent struct {
	offset float64
	event  interface{}
}

type noteOn struct {
	note int
}
type noteOff struct {
	note int
}

// ----- MONO OSC ----- //

type monoOsc struct {
	o           *decoratedOsc
	activeNotes []int
}

func newMonoOsc() *monoOsc {
	return &monoOsc{
		o: &decoratedOsc{
			oscs:      []*osc{{phase: rand.Float64() * 2.0 * math.Pi}, {phase: rand.Float64() * 2.0 * math.Pi}},
			adsr:      &adsr{},
			filter:    &filter{},
			lfos:      []*lfo{newLfo(), newLfo(), newLfo()},
			envelopes: []*envelope{newEnvelope(), newEnvelope(), newEnvelope()},
		},
		activeNotes: make([]int, 0, 128),
	}
}

func (m *monoOsc) calc(
	events [][]*midiEvent,
	oscParams []*oscParams,
	adsrParams *adsrParams,
	filterParams *filterParams,
	lfoParams []*lfoParams,
	envelopeParams []*envelopeParams,
	glideTime int,
	echo *echo,
	out []float64,
) {
	for i, lfo := range m.o.lfos {
		lfo.applyParams(lfoParams[i])
	}
	for i, envelope := range m.o.envelopes {
		envelope.destination = envelopeParams[i].destination
		envelope.adsr.applyEnvelopeParams(envelopeParams[i])
	}
	m.o.adsr.setParams(adsrParams)
	for i := int64(0); i < int64(len(out)); i++ {
		event := enumNoEvent
		for _, e := range events[i] {
			switch data := e.event.(type) {
			case *noteOn:
				if len(m.activeNotes) < cap(m.activeNotes) {
					m.activeNotes = m.activeNotes[:len(m.activeNotes)+1]
					for i := len(m.activeNotes) - 1; i >= 1; i-- {
						m.activeNotes[i] = m.activeNotes[i-1]
					}
					m.activeNotes[0] = data.note
					if len(m.activeNotes) == 1 {
						m.o.initWithNote(oscParams, data.note)
						event = enumNoteOn
					} else {
						m.o.glide(oscParams, m.activeNotes[0], glideTime)
					}
				}
			case *noteOff:
				removed := 0
				for i := 0; i < len(m.activeNotes); i++ {
					if m.activeNotes[i] == data.note {
						removed++
					} else {
						m.activeNotes[i-removed] = m.activeNotes[i]
					}
				}
				m.activeNotes = m.activeNotes[:len(m.activeNotes)-removed]
				if len(m.activeNotes) > 0 {
					m.o.glide(oscParams, m.activeNotes[0], glideTime)
				} else {
					event = enumNoteOff
				}
			}
		}
		out[i] = m.o.step(event, filterParams)
		out[i] = echo.step(out[i])
	}
}

// ----- POLY OSC ----- //

type polyOsc struct {
	// pooled + active = maxPoly
	pooled []*noteOsc
	active []*noteOsc
}

type noteOsc struct {
	*decoratedOsc
	note int
}

func newPolyOsc() *polyOsc {
	pooled := make([]*noteOsc, maxPoly)
	for i := 0; i < len(pooled); i++ {
		pooled[i] = &noteOsc{
			decoratedOsc: &decoratedOsc{
				oscs:      []*osc{{phase: rand.Float64() * 2.0 * math.Pi}, {phase: rand.Float64() * 2.0 * math.Pi}},
				adsr:      &adsr{},
				filter:    &filter{},
				lfos:      []*lfo{newLfo(), newLfo(), newLfo()},
				envelopes: []*envelope{newEnvelope(), newEnvelope(), newEnvelope()},
			},
		}
	}
	return &polyOsc{
		pooled: pooled,
	}
}

func (p *polyOsc) calc(
	events [][]*midiEvent,
	oscParams []*oscParams,
	adsrParams *adsrParams,
	filterParams *filterParams,
	lfoParams []*lfoParams,
	envelopeParams []*envelopeParams,
	echo *echo,
	out []float64,
) {
	for i := int64(0); i < int64(len(out)); i++ {
		events := events[i]
		for j := 0; j < len(events); j++ {
			switch data := events[j].event.(type) {
			case *noteOn:
				lenPooled := len(p.pooled)
				if lenPooled > 0 {
					o := p.pooled[lenPooled-1]
					p.pooled = p.pooled[:lenPooled-1]
					p.active = append(p.active, o)
					o.note = data.note
					o.initWithNote(oscParams, data.note)
					o.adsr.init(adsrParams)
				} else {
					log.Println("maxPoly exceeded")
				}
			}
		}
		for _, o := range p.active {
			for i, lfo := range o.lfos {
				lfo.applyParams(lfoParams[i])
			}
			for i, envelope := range o.envelopes {
				envelope.destination = envelopeParams[i].destination
				envelope.adsr.applyEnvelopeParams(envelopeParams[i])
			}
		}
		out[i] = 0.0
		for _, o := range p.active {
			event := enumNoEvent
			for _, e := range events {
				switch data := e.event.(type) {
				case *noteOff:
					if data.note == o.note {
						event = enumNoteOff
					}
				case *noteOn:
					if data.note == o.note {
						event = enumNoteOn
					}
				}
			}
			out[i] += o.step(event, filterParams)
		}
		for j := len(p.active) - 1; j >= 0; j-- {
			o := p.active[j]
			if o.adsr.phase == phaseNone {
				p.active = append(p.active[:j], p.active[j+1:]...)
				p.pooled = append(p.pooled, o)
			}
		}
		out[i] = echo.step(out[i])
	}
}

// ----- DECORATED OSC ----- //

type decoratedOsc struct {
	oscs      []*osc
	adsr      *adsr
	filter    *filter
	lfos      []*lfo
	envelopes []*envelope
}

const (
	enumNoEvent = iota
	enumNoteOn
	enumNoteOff
)

func (o *decoratedOsc) initWithNote(p []*oscParams, note int) {
	for i, osc := range o.oscs {
		osc.initWithNote(p[i], note)
	}
}
func (o *decoratedOsc) glide(p []*oscParams, note int, glideTime int) {
	for i, osc := range o.oscs {
		osc.glide(p[i], note, glideTime)
	}
}
func (o *decoratedOsc) step(event int, filterParams *filterParams) float64 {
	switch event {
	case enumNoEvent:
	case enumNoteOn:
		o.adsr.noteOn()
		for _, envelope := range o.envelopes {
			envelope.noteOn()
		}
	case enumNoteOff:
		o.adsr.noteOff()
		for _, envelope := range o.envelopes {
			envelope.noteOff()
		}
	}
	o.adsr.step()
	for _, envelope := range o.envelopes {
		if !envelope.enabled {
			continue
		}
		envelope.step()
	}
	freqRatio := 1.0
	phaseShift := 0.0
	ampRatio := 1.0
	filterFreqRatio := 1.0
	for lfoIndex, lfo := range o.lfos {
		amountGain := 1.0
		lfoFreqRatio := 1.0
		for _, envelope := range o.envelopes {
			if !envelope.enabled {
				continue
			}
			if envelope.destination == destLfoAmount[lfoIndex] {
				amountGain *= envelope.value
			}
			if envelope.destination == destLfoFreq[lfoIndex] {
				lfoFreqRatio *= math.Pow(16.0, envelope.value)
			}
		}
		_freqRatio, _phaseShift, _ampRatio, _filterFreqRatio := lfo.step(o.oscs[0], amountGain, lfoFreqRatio) // TODO
		freqRatio *= _freqRatio
		phaseShift += _phaseShift
		ampRatio *= _ampRatio
		filterFreqRatio *= _filterFreqRatio
	}
	for _, envelope := range o.envelopes {
		if !envelope.enabled {
			continue
		}
		if envelope.destination == destFreq {
			freqRatio *= math.Pow(16.0, envelope.value)
		}
	}
	value := 0.0
	for _, osc := range o.oscs {
		value += osc.step(freqRatio, phaseShift) * oscGain * ampRatio * o.adsr.value
	}
	// TODO
	o.filter.enabled = filterParams.enabled
	o.filter.kind = filterParams.kind
	o.filter.freq = filterParams.freq
	o.filter.q = filterParams.q
	o.filter.gain = filterParams.gain
	o.filter.N = filterParams.N
	return o.filter.step(value, filterFreqRatio, o.envelopes)
}

// ----- OSC ----- //

type oscParams struct {
	enabled bool
	kind    int
	octave  int     // -2 ~ 2
	coarse  int     // -12 ~ 12
	fine    int     // -100 ~ 100 cent
	level   float64 // 0 ~ 1
}
type oscJSON struct {
	Enabled bool    `json:"enabled"`
	Kind    string  `json:"kind"`
	Octave  int     `json:"octave"`
	Coarse  int     `json:"coarse"`
	Fine    int     `json:"fine"`
	Level   float64 `json:"level"`
}

func (o *oscParams) applyJSON(data json.RawMessage) {
	var j oscJSON
	err := json.Unmarshal(data, &j)
	if err != nil {
		log.Println("failed to apply JSON to oscParams")
		return
	}
	o.enabled = j.Enabled
	o.kind = waveKindFromString(j.Kind)
	o.octave = j.Octave
	o.coarse = j.Coarse
	o.fine = j.Fine
	o.level = j.Level
}
func (o *oscParams) toJSON() json.RawMessage {
	return toRawMessage(&oscJSON{
		Enabled: o.enabled,
		Kind:    waveKindToString(o.kind),
		Octave:  o.octave,
		Coarse:  o.coarse,
		Fine:    o.fine,
		Level:   o.level,
	})
}
func (o *oscParams) set(key string, value string) error {
	switch key {
	case "enabled":
		o.enabled = value == "true"
	case "kind":
		o.kind = waveKindFromString(value)
	case "octave":
		value, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		o.octave = int(value)
	case "coarse":
		value, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		o.coarse = int(value)
	case "fine":
		value, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		o.fine = int(value)
	case "level":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		o.level = value
	}
	return nil
}

type osc struct {
	enabled   bool
	kind      int
	glideTime int // ms
	freq      float64
	level     float64
	phase     float64
	gliding   bool
	shiftPos  float64
	prevFreq  float64
	nextFreq  float64
}

var blsquareWT *WavetableSet = loadWavetableSet("work/square")
var blsawWT *WavetableSet = loadWavetableSet("work/saw")

func loadWavetableSet(path string) *WavetableSet {
	wts := NewWavetableSet(128, 4096)
	wts.Load(path)
	return wts
}
func noteWithParamsToFreq(p *oscParams, note int) float64 {
	return noteToFreq(note) * math.Pow(2, float64(p.octave)+float64(p.coarse)/12+float64(p.fine)/100/12)
}
func (o *osc) initWithNote(p *oscParams, note int) {
	o.enabled = p.enabled
	o.kind = p.kind
	o.freq = noteWithParamsToFreq(p, note)
	o.level = p.level
	o.phase = rand.Float64() * 2.0 * math.Pi
}
func (o *osc) glide(p *oscParams, note int, glideTime int) {
	nextFreq := noteWithParamsToFreq(p, note)
	if math.Abs(nextFreq-o.freq) < 0.001 {
		return
	}
	o.enabled = p.enabled
	o.glideTime = glideTime
	o.prevFreq = o.freq
	o.nextFreq = nextFreq
	o.gliding = true
	o.shiftPos = 0
}
func (o *osc) step(freqRatio float64, phaseShift float64) float64 {
	if !o.enabled {
		return 0.0
	}
	freq := o.freq * freqRatio
	phase := o.phase + phaseShift
	value := 0.0
	switch o.kind {
	case waveSine:
		value = math.Sin(phase)
	case waveTriangle:
		p := positiveMod(phase/(2.0*math.Pi), 1)
		if p < 0.5 {
			value = p*4 - 1
		} else {
			value = p*(-4) + 3
		}
	case waveSquare:
		p := positiveMod(phase/(2.0*math.Pi), 1)
		if p < 0.5 {
			value = 1
		} else {
			value = -1
		}
	case waveSquareWT:
		note := freqToNote(freq)
		value = blsquareWT.tables[note].getAtPhase(phase)
	case wavePulse:
		p := positiveMod(phase/(2.0*math.Pi), 1)
		if p < 0.25 {
			value = 1
		} else {
			value = -1
		}
	case waveSaw:
		p := positiveMod(phase/(2.0*math.Pi), 1)
		value = p*2 - 1
	case waveSawWT:
		note := freqToNote(freq)
		value = blsawWT.tables[note].getAtPhase(phase)
	case waveSawRev:
		p := positiveMod(phase/(2.0*math.Pi), 1)
		value = p*(-2) + 1
	case waveNoise:
		value = rand.Float64()*2 - 1
	}
	o.phase += 2.0 * math.Pi * freq / float64(sampleRate)
	if o.gliding {
		o.shiftPos++
		t := o.shiftPos * secPerSample * 1000 / float64(o.glideTime)
		o.freq = t*o.nextFreq + (1-t)*o.prevFreq
		if t >= 1 || math.Abs(o.nextFreq-o.freq) < 0.001 {
			o.freq = o.nextFreq
			o.gliding = false
		}
	}
	return value * o.level
}

// ----- Wave Kind ----- //

/*
generate-enum waveKind

waveNone none
waveSine sine
waveTriangle triangle
waveSquare square
waveSquareWT square-wt
wavePulse pulse
waveSaw saw
waveSawWT saw-wt
waveSawRev saw-rev
waveNoise noise

EOF
*/

// ----- ADSR ----- //

const (
	phaseNone = iota
	phaseAttack
	phaseKeep
	phaseDecay
	phaseSustain
	phaseRelease
)

type adsrParams struct {
	attack  float64 // ms
	decay   float64 // ms
	sustain float64 // 0-1
	release float64 // ms
}
type adsrJSON struct {
	Attack  float64 `json:"attack"`
	Decay   float64 `json:"decay"`
	Sustain float64 `json:"sustain"`
	Release float64 `json:"release"`
}

func (a *adsrParams) applyJSON(data json.RawMessage) {
	var j adsrJSON
	err := json.Unmarshal(data, &j)
	if err != nil {
		log.Println("failed to apply JSON to adsrParams")
		return
	}
	a.attack = j.Attack
	a.decay = j.Decay
	a.sustain = j.Sustain
	a.release = j.Release
}
func (a *adsrParams) toJSON() json.RawMessage {
	return toRawMessage(&adsrJSON{
		Attack:  a.attack,
		Decay:   a.decay,
		Sustain: a.sustain,
		Release: a.release,
	})
}
func (a *adsrParams) set(key string, value string) error {
	switch key {
	case "attack":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		a.attack = value
	case "decay":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		a.decay = value
	case "sustain":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		a.sustain = value
	case "release":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		a.release = value
	}
	return nil
}

/*
  p +     x---x
    |    /     \
    |   /       \
  s +  /         x------x
    | /                  \
    |/                    \
  b +-----+---+--+------+---
    |a    |k  |d |      |r |
*/
type adsr struct {
	attack         float64 // ms
	keep           float64 // ms
	decay          float64 // ms
	sustain        float64 // 0-1
	release        float64 // ms
	base           float64 // 0-1
	peek           float64 // 0-1
	value          float64
	phase          int // "attack", "keep", "decay", "sustain", "release" nil
	phasePos       int
	valueAtNoteOn  float64
	valueAtNoteOff float64
}

func (a *adsr) init(p *adsrParams) {
	a.setParams(p)
	a.value = 0
	a.phase = phaseNone
	a.phasePos = 0
	a.valueAtNoteOn = 0
	a.valueAtNoteOff = 0
}

func (a *adsr) setParams(p *adsrParams) {
	a.base = 0
	a.peek = 1
	a.attack = p.attack
	a.keep = 0
	a.decay = p.decay
	a.sustain = p.sustain
	a.release = p.release
}
func (a *adsr) applyEnvelopeParams(p *envelopeParams) {
	if p.destination == destVibrato ||
		p.destination == destTremolo ||
		p.destination == destFilterQ0V ||
		p.destination == destFilterGain0V ||
		p.destination == destLfo0Amount ||
		p.destination == destLfo1Amount ||
		p.destination == destLfo2Amount {
		// zero-to-value
		a.base = 1
		a.peek = 0
	} else if p.destination == destFreq ||
		p.destination == destFilterFreq ||
		p.destination == destLfo0Freq ||
		p.destination == destLfo1Freq ||
		p.destination == destLfo2Freq {
		// amount-to-value
		a.base = 0
		a.peek = p.amount
	} else if p.destination == destFilterQ ||
		p.destination == destFilterGain {
		// value-to-zero
		a.base = 0
		a.peek = 1
	}
	a.attack = 0
	a.keep = p.delay
	a.decay = p.attack
	a.sustain = a.base
	a.release = 0
}

func (a *adsr) noteOn() {
	a.phase = phaseAttack
	a.phasePos = 0
	a.valueAtNoteOn = a.value
}

func (a *adsr) noteOff() {
	a.phase = phaseRelease
	a.phasePos = 0
	a.valueAtNoteOff = a.value
}

func (a *adsr) step() {
	phaseTime := float64(a.phasePos) * secPerSample * 1000 // ms
	switch a.phase {
	case phaseAttack:
		if phaseTime >= float64(a.attack) {
			a.phase = phaseKeep
			a.phasePos = 0
			a.value = a.peek
		} else {
			t := phaseTime / float64(a.attack)
			a.value = t*a.peek + (1-t)*a.valueAtNoteOn // TODO: don't use the same attack time
			a.phasePos++
		}
	case phaseKeep:
		if phaseTime >= float64(a.keep) {
			a.phase = phaseDecay
			a.phasePos = 0
		} else {
			a.phasePos++
		}
	case phaseDecay:
		ended := false
		if a.decay == 0 {
			ended = true
		} else {
			t := phaseTime / float64(a.decay)
			a.value = setTargetAtTime(a.peek, a.sustain, t)
			if math.Abs(a.value-a.sustain) < 0.001 {
				ended = true
			}
		}
		if ended {
			a.phase = phaseSustain
			a.phasePos = 0
			a.value = float64(a.sustain)
		} else {
			a.phasePos++
		}
	case phaseSustain:
		a.value = float64(a.sustain)
	case phaseRelease:
		ended := false
		if a.release == 0 {
			ended = true
		} else {
			t := phaseTime / float64(a.release)
			a.value = setTargetAtTime(a.valueAtNoteOff, a.base, t)
			if math.Abs(a.value-a.base) < 0.001 {
				ended = true
			}
		}
		if ended {
			a.phase = phaseNone
			a.phasePos = 0
			a.value = a.base
		} else {
			a.phasePos++
		}
	default:
		a.value = a.base
	}
}

// 63% closer to target when pos=1.0
func setTargetAtTime(initialValue float64, targetValue float64, pos float64) float64 {
	return targetValue + (initialValue-targetValue)*math.Exp(-pos)
}

// ----- Filter Kind ----- //

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

// ----- Filter ----- //

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
			freqRatio *= math.Pow(16.0, envelope.value)
		}
		if envelope.destination == destFilterQ || envelope.destination == destFilterQ0V {
			qExponent *= envelope.value
		}
		if envelope.destination == destFilterGain || envelope.destination == destFilterGain0V {
			gainRatio *= envelope.value
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
func calcFilter(in []float64, out []float64, a []float64, b []float64, past []float64) {
	for i := 0; i < len(in); i++ {
		out[i] = calcFilterOneSample(in[i], a, b, past)
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
func impulseResponse(a []float64, b []float64, n int) []float64 {
	in := make([]float64, n)
	out := make([]float64, n)
	past := make([]float64, int(math.Max(float64(len(a)-1), float64(len(b)))))
	in[0] = 1
	calcFilter(in, out, a, b, past)
	return out
}
func frequencyResponse(a []float64, b []float64) []float64 {
	h := impulseResponse(a, b, fftSize)
	fft.CalcAbs(h)
	return h[:fftSize/2]
}

// ----- Delay ----- //

type delay struct {
	cursor int
	past   []float64
}

func (d *delay) applyParams(millis float64) {
	if millis < 10 {
		millis = 10
	}
	length := int(sampleRate * millis / 1000)
	if cap(d.past) >= length {
		d.past = d.past[0:length]
	} else {
		d.past = make([]float64, length)
	}
	if d.cursor >= len(d.past) {
		d.cursor = 0
	}
}

func (d *delay) step(in float64) {
	d.past[d.cursor] = in
	d.cursor++
	if d.cursor >= len(d.past) {
		d.cursor = 0
	}
}
func (d *delay) getDelayed() float64 {
	return d.past[d.cursor]
}

// ----- Echo ----- //

type echoParams struct {
	enabled      bool
	delay        float64
	feedbackGain float64
	mix          float64
}

type echoJSON struct {
	Enabled      bool    `json:"enabled"`
	Delay        float64 `json:"delay"`
	FeedbackGain float64 `json:"feedbackGain"`
	Mix          float64 `json:"mix"`
}

func (l *echoParams) applyJSON(data json.RawMessage) {
	var j echoJSON
	err := json.Unmarshal(data, &j)
	if err != nil {
		log.Println("failed to apply JSON to echoParams")
		return
	}
	l.enabled = j.Enabled
	l.delay = j.Delay
	l.feedbackGain = j.FeedbackGain
	l.mix = j.Mix
}
func (l *echoParams) toJSON() json.RawMessage {
	return toRawMessage(&echoJSON{
		Enabled:      l.enabled,
		Delay:        l.delay,
		FeedbackGain: l.feedbackGain,
		Mix:          l.mix,
	})
}
func (l *echoParams) set(key string, value string) error {
	switch key {
	case "enabled":
		l.enabled = value == "true"
	case "delay":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		l.delay = value
	case "feedbackGain":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		l.feedbackGain = value
	case "mix":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		l.mix = value
	}
	return nil
}

type echo struct {
	enabled      bool
	delay        *delay
	feedbackGain float64 // [0,1)
	mix          float64 // [0,1]
}

func (e *echo) applyParams(p *echoParams) {
	e.enabled = p.enabled
	e.delay.applyParams(p.delay)
	e.feedbackGain = p.feedbackGain
	e.mix = p.mix
}

func (e *echo) step(in float64) float64 {
	if !e.enabled {
		return in
	}
	delayed := e.delay.getDelayed()
	e.delay.step(in + delayed*e.feedbackGain)
	return in + delayed*e.mix
}

// ----- Destination ----- //

/*
generate-enum destination

destNone none
destVibrato vibrato
destTremolo tremolo
destFM fm
destPM pm
destAM am
destFreq freq
destFilterFreq filter_freq
destFilterQ filter_q
destFilterQ0V filter_q_0v
destFilterGain filter_gain
destFilterGain0V filter_gain_0v
destLfo0Freq lfo0_freq
destLfo1Freq lfo1_freq
destLfo2Freq lfo2_freq
destLfo0Amount lfo0_amount
destLfo1Amount lfo1_amount
destLfo2Amount lfo2_amount

EOF
*/

var destLfoFreq = [3]int{destLfo0Freq, destLfo1Freq, destLfo2Freq}
var destLfoAmount = [3]int{destLfo0Amount, destLfo1Amount, destLfo2Amount}

// ----- Envelope ----- //

type envelopeParams struct {
	enabled     bool
	destination int
	delay       float64
	attack      float64
	amount      float64
}

type envelopeJSON struct {
	Enabled     bool    `json:"enabled"`
	Destination string  `json:"destination"`
	Delay       float64 `json:"delay"`
	Attack      float64 `json:"attack"`
	Amount      float64 `json:"amount"`
}

func (l *envelopeParams) applyJSON(data json.RawMessage) {
	var j envelopeJSON
	err := json.Unmarshal(data, &j)
	if err != nil {
		log.Println("failed to apply JSON to envelopeParams")
		return
	}
	l.enabled = j.Enabled
	l.destination = destinationFromString(j.Destination)
	l.delay = j.Delay
	l.attack = j.Attack
	l.amount = j.Amount
}
func (l *envelopeParams) toJSON() json.RawMessage {
	return toRawMessage(&envelopeJSON{
		Enabled:     l.enabled,
		Destination: destinationToString(l.destination),
		Delay:       l.delay,
		Attack:      l.attack,
		Amount:      l.amount,
	})
}
func newEnvelopeParams() *envelopeParams {
	return &envelopeParams{
		enabled:     false,
		destination: destNone,
		delay:       0,
		attack:      0,
		amount:      0,
	}
}
func (l *envelopeParams) set(key string, value string) error {
	switch key {
	case "enabled":
		l.enabled = value == "true"
	case "destination":
		l.destination = destinationFromString(value)
	case "delay":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		l.delay = value
	case "attack":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		l.attack = value
	case "amount":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		l.amount = value
	}
	return nil
}

/*
  [value-to-zero] filter_q, etc.
  v +_
	  | \_
    |   \_
    |     \_
  0 +-------+---------
    |attack |

  [amount-to-value] freq, filter_freq, etc.
  a +_
	  | \_
    |   \_
    |     \_
  v +       `-----
	  |
		+--------+--------
    |attack  |

  [zero-to-value] vibrato, tremolo, etc.
  v +           ,-----
	  |         _/
    |       _/
    |      /
  0 +-----+------+----
    |delay|attack|
*/
type envelope struct {
	enabled bool
	*adsr
	destination int
}

func newEnvelope() *envelope {
	return &envelope{
		enabled:     false,
		destination: destNone,
		adsr:        &adsr{},
	}
}

// ----- LFO ----- //

type lfoParams struct {
	enabled     bool
	destination int
	wave        int
	freqType    string
	freq        float64
	amount      float64
}

type lfoJSON struct {
	Enabled     bool    `json:"enabled"`
	Destination string  `json:"destination"`
	Wave        string  `json:"wave"`
	FreqType    string  `json:"freqType"`
	Freq        float64 `json:"freq"`
	Amount      float64 `json:"amount"`
}

func (l *lfoParams) applyJSON(data json.RawMessage) {
	var j lfoJSON
	err := json.Unmarshal(data, &j)
	if err != nil {
		log.Println("failed to apply JSON to adsrParams")
		return
	}
	l.enabled = j.Enabled
	l.destination = destinationFromString(j.Destination)
	l.wave = waveKindFromString(j.Wave)
	l.freqType = j.FreqType
	l.freq = j.Freq
	l.amount = j.Amount
}
func (l *lfoParams) toJSON() json.RawMessage {
	return toRawMessage(&lfoJSON{
		Enabled:     l.enabled,
		Destination: destinationToString(l.destination),
		Wave:        waveKindToString(l.wave),
		FreqType:    l.freqType,
		Freq:        l.freq,
		Amount:      l.amount,
	})
}

func newLfoParams() *lfoParams {
	return &lfoParams{
		enabled:     false,
		destination: destNone,
		wave:        waveSine,
		freqType:    "none",
		freq:        0,
		amount:      0,
	}
}

// func (l *lfoParams) initByDestination(destination string) {
// 	l.destination = destination
// 	switch destination {
// 	case "vibrato":
// 		l.freqType = "absolute"
// 		l.freq = 0
// 		l.amount = 0
// 	case "tremolo":
// 		l.freqType = "absolute"
// 		l.freq = 0
// 		l.amount = 0
// 	}
// }

func (l *lfoParams) set(key string, value string) error {
	switch key {
	case "enabled":
		// l.initByDestination(value)
		l.enabled = value == "true"
	case "destination":
		// l.initByDestination(value)
		l.destination = destinationFromString(value)
	case "wave":
		l.wave = waveKindFromString(value)
	case "freq":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		l.freq = value
	case "amount":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		l.amount = value
	}
	return nil
}

type lfo struct {
	enabled     bool
	destination int
	freqType    string
	amount      float64
	osc         *osc
}

func newLfo() *lfo {
	return &lfo{
		enabled:     false,
		destination: destNone,
		freqType:    "none",
		amount:      0,
		osc:         &osc{phase: rand.Float64() * 2.0 * math.Pi, level: 1.0},
	}
}

func (l *lfo) applyParams(p *lfoParams) {
	l.enabled = p.enabled
	l.destination = p.destination
	l.osc.kind = p.wave
	l.freqType = p.freqType
	l.osc.freq = p.freq
	l.amount = p.amount
}

func (l *lfo) step(career *osc, amountGain float64, lfoFreqRatio float64) (float64, float64, float64, float64) {
	freqRatio := 1.0
	phaseShift := 0.0
	ampRatio := 1.0
	filterFreqRatio := 1.0
	if !l.enabled {
		return freqRatio, phaseShift, ampRatio, filterFreqRatio
	}
	switch l.destination {
	case destVibrato:
		amount := l.amount * amountGain
		freqRatio = math.Pow(2.0, l.osc.step(lfoFreqRatio, 0.0)*amount/100.0/12.0)
	case destTremolo:
		amount := l.amount * amountGain
		ampRatio = 1.0 + (l.osc.step(lfoFreqRatio, 0.0)-1.0)/2.0*amount
	case destFM:
		amount := l.amount * amountGain
		freqRatio = math.Pow(2.0, l.osc.step(career.freq*lfoFreqRatio, 0.0)*amount/100/12)
	case destPM:
		amount := l.amount * amountGain
		phaseShift = l.osc.step(career.freq*lfoFreqRatio, 0.0) * amount
	case destAM:
		amount := l.amount * amountGain
		ampRatio = 1.0 + l.osc.step(career.freq*lfoFreqRatio, 0.0)*amount
	case destFilterFreq:
		amount := l.amount * amountGain
		filterFreqRatio = math.Pow(16.0, l.osc.step(lfoFreqRatio, 0.0)*amount)
	}
	return freqRatio, phaseShift, ampRatio, filterFreqRatio
}

// ----- Changes ----- //

// Changes ...
type Changes struct {
	sync.Mutex
	dict map[string]struct{}
}

// Add ...
func (c *Changes) Add(key string) {
	c.Lock()
	c.dict[key] = struct{}{}
	c.Unlock()
}

// Has ...
func (c *Changes) Has(key string) bool {
	c.Lock()
	_, ok := c.dict[key]
	c.Unlock()
	return ok
}

// Delete ...
func (c *Changes) Delete(key string) {
	c.Lock()
	delete(c.dict, key)
	c.Unlock()
}

// ----- State ----- //

type state struct {
	sync.Mutex
	events         [][]*midiEvent // length: samplesPerCycle * 2
	oscParams      []*oscParams
	adsrParams     *adsrParams
	lfoParams      []*lfoParams
	filterParams   *filterParams
	envelopeParams []*envelopeParams
	echoParams     *echoParams
	polyMode       bool
	glideTime      int // ms
	monoOsc        *monoOsc
	polyOsc        *polyOsc
	echo           *echo
	pos            int64
	out            []float64 // length: fftSize
	lastRead       float64
	processTime    float64
}
type stateJSON struct {
	Poly      string            `json:"poly"`
	GlideTime int               `json:"glideTime"`
	Oscs      []json.RawMessage `json:"oscs"`
	Adsr      json.RawMessage   `json:"adsr"`
	Lfos      []json.RawMessage `json:"lfos"`
	Envelopes []json.RawMessage `json:"envelopes"`
	Filter    json.RawMessage   `json:"filter"`
	Echo      json.RawMessage   `json:"echo"`
}

func (s *state) applyJSON(data json.RawMessage) {
	var j stateJSON
	err := json.Unmarshal(data, &j)
	if err != nil {
		log.Println(err)
		log.Println(data)
		log.Println("failed to apply JSON to state")
		return
	}
	s.polyMode = j.Poly == "poly"
	s.glideTime = j.GlideTime
	if len(j.Oscs) == len(s.oscParams) {
		for i, j := range j.Oscs {
			s.oscParams[i].applyJSON(j)
		}
	} else {
		log.Println("failed to apply JSON to osc params")
	}
	s.adsrParams.applyJSON(j.Adsr)
	s.filterParams.applyJSON(j.Filter)
	if len(j.Lfos) == len(s.lfoParams) {
		for i, j := range j.Lfos {
			s.lfoParams[i].applyJSON(j)
		}
	} else {
		log.Println("failed to apply JSON to lfo params")
	}
	if len(j.Envelopes) == len(s.envelopeParams) {
		for i, j := range j.Envelopes {
			s.envelopeParams[i].applyJSON(j)
		}
	} else {
		log.Println("failed to apply JSON to envelope params")
	}
	s.echoParams.applyJSON(j.Echo)
}
func (s *state) toJSON() json.RawMessage {
	oscJsons := make([]json.RawMessage, len(s.oscParams))
	for i, oscParam := range s.oscParams {
		oscJsons[i] = oscParam.toJSON()
	}
	lfoJsons := make([]json.RawMessage, len(s.lfoParams))
	for i, lfoParam := range s.lfoParams {
		lfoJsons[i] = lfoParam.toJSON()
	}
	envelopeJsons := make([]json.RawMessage, len(s.envelopeParams))
	for i, envelopeParam := range s.envelopeParams {
		envelopeJsons[i] = envelopeParam.toJSON()
	}
	poly := "mono"
	if s.polyMode {
		poly = "poly"
	}
	return toRawMessage(&stateJSON{
		Poly:      poly,
		GlideTime: s.glideTime,
		Oscs:      oscJsons,
		Adsr:      s.adsrParams.toJSON(),
		Lfos:      lfoJsons,
		Envelopes: envelopeJsons,
		Filter:    s.filterParams.toJSON(),
		Echo:      s.echoParams.toJSON(),
	})
}

func newState() *state {
	return &state{
		events:         make([][]*midiEvent, samplesPerCycle*2),
		oscParams:      []*oscParams{{kind: waveSine, level: 1.0}, {kind: waveSine, level: 1.0}},
		adsrParams:     &adsrParams{attack: 10, decay: 100, sustain: 0.7, release: 200},
		lfoParams:      []*lfoParams{newLfoParams(), newLfoParams(), newLfoParams()},
		filterParams:   &filterParams{kind: filterNone, freq: 1000, q: 1, gain: 0, N: 50},
		envelopeParams: []*envelopeParams{newEnvelopeParams(), newEnvelopeParams(), newEnvelopeParams()},
		echoParams:     &echoParams{},
		polyMode:       false,
		glideTime:      100,
		monoOsc:        newMonoOsc(),
		polyOsc:        newPolyOsc(),
		echo:           &echo{delay: &delay{}},
		pos:            0,
		out:            make([]float64, fftSize),
	}
}

// ----- Audio ----- //

// Audio ...
type Audio struct {
	ctx        context.Context
	otoContext *oto.Context
	CommandCh  chan []string
	state      *state
	Changes    *Changes
	fftResult  []float64 // length: fftSize
}

var _ io.Reader = (*Audio)(nil)

type audioJSON struct {
	State json.RawMessage `json:"state"`
	// additional context...
}

// ApplyJSON ...
func (a *Audio) ApplyJSON(data []byte) {
	a.state.Lock()
	defer a.state.Unlock()
	var audioJSON audioJSON
	err := json.Unmarshal(data, &audioJSON)
	if err != nil {
		log.Println("failed to apply JSON to Audio", err)
		return
	}
	a.state.applyJSON(audioJSON.State)
}

// ToJSON ...
func (a *Audio) ToJSON() []byte {
	a.state.Lock()
	defer a.state.Unlock()
	bytes, err := json.Marshal(a.toJSON())
	if err != nil {
		panic(err)
	}
	return bytes
}

func (a *Audio) toJSON() json.RawMessage {
	return toRawMessage(&audioJSON{
		State: a.state.toJSON(),
	})
}

func (a *Audio) Read(buf []byte) (int, error) {
	select {
	case <-a.ctx.Done():
		log.Println("Read() interrupted.")
		return 0, io.EOF
	default:
		a.state.Lock()
		defer a.state.Unlock()
		timestamp := now()
		bufSamples := int64(len(buf) / bytesPerSample)

		offset := a.state.pos % fftSize
		out := a.state.out[offset : offset+bufSamples]

		a.state.echo.applyParams(a.state.echoParams)
		if a.state.polyMode {
			a.state.polyOsc.calc(a.state.events, a.state.oscParams, a.state.adsrParams, a.state.filterParams, a.state.lfoParams, a.state.envelopeParams, a.state.echo, out)
		} else {
			a.state.monoOsc.calc(a.state.events, a.state.oscParams, a.state.adsrParams, a.state.filterParams, a.state.lfoParams, a.state.envelopeParams, a.state.glideTime, a.state.echo, out)
		}
		writeBuffer(a.state.out, offset, buf, 0)
		writeBuffer(a.state.out, offset, buf, 1)
		a.state.pos += bufSamples
		a.state.lastRead = timestamp
		eventLength := len(a.state.events)
		for i := 0; i < eventLength; i++ {
			if i >= eventLength/2 {
				a.state.events[i-eventLength/2] = a.state.events[i]
			}
			a.state.events[i] = nil
		}
		endTime := now()
		a.state.processTime = endTime - timestamp
		if a.state.processTime > responseDelay {
			log.Printf("[WARN] time budget exceeded: processTime=%dms, activeNotes=%d\n",
				int(a.state.processTime*1000), len(a.state.polyOsc.active))
		} else {
			// log.Printf("%.2fms\n", a.state.processTime*1000)
		}
		// log.Println(len(a.state.polyOsc.active))
		return len(buf), nil // io.EOF, etc.
	}
}

func writeBuffer(out []float64, outOffset int64, buf []byte, ch int) {
	sampleLength := int(len(buf) / bytesPerSample)
	for i := 0; i < sampleLength; i++ {
		value := out[outOffset+int64(i)]
		switch bitDepthInBytes {
		case 1:
			const max = 127
			b := int(value * max)
			buf[bytesPerSample*i+ch] = byte(b + 128)
		case 2:
			const max = 32767
			b := int16(value * max)
			buf[bytesPerSample*i+2*ch] = byte(b)
			buf[bytesPerSample*i+2*ch+1] = byte(b >> 8)
		}
	}
}

// NewAudio ...
func NewAudio() (*Audio, error) {
	otoContext, err := oto.NewContext(sampleRate, channelNum, bitDepthInBytes, bufferSizeInBytes)
	if err != nil {
		return nil, err
	}
	commandCh := make(chan []string, 256)
	audio := &Audio{
		ctx:        context.Background(),
		otoContext: otoContext,
		CommandCh:  commandCh,
		state:      newState(),
		Changes: &Changes{
			dict: make(map[string]struct{}),
		},
		fftResult: make([]float64, fftSize),
	}
	go processCommands(audio, commandCh)
	return audio, nil
}

func processCommands(audio *Audio, commandCh <-chan []string) {
	for command := range commandCh {
		err := audio.update(command)
		if err != nil {
			panic(err)
		}
	}
	log.Println("processCommands() ended.")
}

func (a *Audio) update(command []string) error {
	a.state.Lock()
	defer a.state.Unlock()

	switch command[0] {
	case "set":
		command = command[1:]
		switch command[0] {
		case "glide_time":
			command = command[1:]
			value, err := strconv.ParseInt(command[0], 10, 64)
			if err != nil {
				return err
			}
			a.state.glideTime = int(value)
		case "osc":
			command = command[1:]
			index, err := strconv.ParseInt(command[0], 10, 64)
			if err != nil {
				return err
			}
			command = command[1:]
			if len(command) != 2 {
				return fmt.Errorf("invalid key-value pair %v", command)
			}
			err = a.state.oscParams[index].set(command[0], command[1])
			if err != nil {
				return err
			}
		case "adsr":
			command = command[1:]
			if len(command) != 2 {
				return fmt.Errorf("invalid key-value pair %v", command)
			}
			err := a.state.adsrParams.set(command[0], command[1])
			if err != nil {
				return err
			}
		case "filter":
			command = command[1:]
			if len(command) != 2 {
				return fmt.Errorf("invalid key-value pair %v", command)
			}
			err := a.state.filterParams.set(command[0], command[1])
			if err != nil {
				return err
			}
			a.Changes.Add("filter-shape")
		case "lfo":
			command = command[1:]
			index, err := strconv.ParseInt(command[0], 10, 64)
			if err != nil {
				return err
			}
			command = command[1:]
			if len(command) != 2 {
				return fmt.Errorf("invalid key-value pair %v", command)
			}
			err = a.state.lfoParams[index].set(command[0], command[1])
			if err != nil {
				return err
			}
		case "envelope":
			command = command[1:]
			index, err := strconv.ParseInt(command[0], 10, 64)
			if err != nil {
				return err
			}
			command = command[1:]
			if len(command) != 2 {
				return fmt.Errorf("invalid key-value pair %v", command)
			}
			err = a.state.envelopeParams[index].set(command[0], command[1])
			if err != nil {
				return err
			}
		case "echo":
			command = command[1:]
			if len(command) != 2 {
				return fmt.Errorf("invalid key-value pair %v", command)
			}
			err := a.state.echoParams.set(command[0], command[1])
			if err != nil {
				return err
			}
		}
		a.Changes.Add("data")
	case "mono":
		a.state.polyMode = false
		a.Changes.Add("data")
	case "poly":
		a.state.polyMode = true
		a.Changes.Add("data")
	case "note_on":
		note, err := strconv.ParseInt(command[1], 10, 32)
		if err != nil {
			return err
		}
		a.addMidiEvent(&noteOn{note: int(note)})
	case "note_off":
		note, err := strconv.ParseInt(command[1], 10, 32)
		if err != nil {
			return err
		}
		a.addMidiEvent(&noteOff{note: int(note)})
	default:
		return fmt.Errorf("unknown command %v", command[0])
	}
	return nil
}

// Close ...
func (a *Audio) Close() error {
	log.Println("Closing Audio...")
	close(a.CommandCh)
	return a.otoContext.Close()
}

// Start ...
func (a *Audio) Start(ctx context.Context) error {
	p := a.otoContext.NewPlayer()
	defer func() {
		if err := p.Close(); err != nil {
			log.Printf("error: %v", err)
		}
	}()
	a.ctx = ctx

	// block until cancel() called
	if _, err := io.CopyBuffer(p, a, make([]byte, bufferSizeInBytes)); err != nil {
		return err
	}
	log.Println("Start() ended.")
	return nil
}

// GetParamsJSON ...
func (a *Audio) GetParamsJSON() []byte {
	a.state.Lock()
	defer a.state.Unlock()
	bytes, err := json.Marshal(a.state.toJSON())
	if err != nil {
		panic(err)
	}
	return bytes
}

var filterShapeFeedforward = []float64{}
var filterShapeFeedback = []float64{}

// GetFilterShape ...
func (a *Audio) GetFilterShape() []float64 {
	a.state.Lock()
	kind := a.state.filterParams.kind
	if !a.state.filterParams.enabled {
		kind = filterNone
	}
	filterShapeFeedforward, filterShapeFeedback := makeH(
		filterShapeFeedforward,
		filterShapeFeedback,
		kind,
		a.state.filterParams.N,
		a.state.filterParams.freq,
		a.state.filterParams.q,
		a.state.filterParams.gain,
	)
	a.state.Unlock()
	return frequencyResponse(filterShapeFeedforward, filterShapeFeedback)
}

type statusJSON struct {
	Polyphony   int     `json:"polyphony"`
	ProcessTime float64 `json:"processTime"`
}

// GetStatusJSON ...
func (a *Audio) GetStatusJSON() []byte {
	a.state.Lock()
	statusJSON := &statusJSON{
		Polyphony:   len(a.state.polyOsc.active),
		ProcessTime: a.state.processTime,
	}
	a.state.Unlock()
	bytes, err := json.Marshal(statusJSON)
	if err != nil {
		panic(err)
	}
	return bytes
}

// GetFFT ...
func (a *Audio) GetFFT() []float64 {
	a.state.Lock()
	// out:       | 4 | 1 | 2 | 3 |
	// offset:        ^
	// fftResult: | 1 | 2 | 3 | 4 |
	// return:    |<----->|
	offset := a.state.pos % fftSize
	copy(a.fftResult, a.state.out[offset:])
	copy(a.fftResult[fftSize-offset:], a.state.out[:offset])
	a.state.Unlock()
	applyWindow(a.fftResult, han)
	fft.CalcAbs(a.fftResult)
	for i, value := range a.fftResult {
		a.fftResult[i] = value * 2 / fftSize
	}
	return a.fftResult[:fftSize/2]
}

// AddMidiEvent ...
func (a *Audio) AddMidiEvent(data []byte) {
	a.state.Lock()
	defer a.state.Unlock()
	if data[0]>>4 == 8 || data[0]>>4 == 9 && data[2] == 0 {
		log.Printf("got note-off: %v\n", data)
		note := int(data[1])
		a.addMidiEvent(&noteOff{note: note})
	} else if data[0]>>4 == 9 && data[2] > 0 {
		log.Printf("got note-on: %v\n", data)
		note := int(data[1])
		a.addMidiEvent(&noteOn{note: note})
	}
}

func (a *Audio) addMidiEvent(event interface{}) {
	offset := now() - a.state.lastRead
	index := int(offset / secPerSample)
	if index < 0 {
		log.Println("[WARN] index < 0")
		index = 0
	}
	if index >= len(a.state.events) {
		log.Println("[WARN] index >= event length")
		index = len(a.state.events) - 1
	}
	a.state.events[index] = append(a.state.events[index], &midiEvent{offset: offset, event: event})
}
