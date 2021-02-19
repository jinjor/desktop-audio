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
func freqToNote(freq float64) int {
	note := int(math.Log2(freq/baseFreq)*12.0) + 69
	if note < 0 {
		note = 0
	}
	if note >= 128 {
		note = 127
	}
	return note
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
	osc         *osc
	adsr        *adsr
	lfos        []*lfo
	envelopes   []*envelope
	activeNotes []int
	out         []float64 // length: samplesPerCycle
}

func newMonoOsc() *monoOsc {
	return &monoOsc{
		osc:         &osc{phase01: rand.Float64()},
		adsr:        &adsr{},
		lfos:        []*lfo{newLfo(), newLfo(), newLfo()},
		envelopes:   []*envelope{newEnvelope(), newEnvelope(), newEnvelope()},
		activeNotes: make([]int, 0, 128),
		out:         make([]float64, samplesPerCycle),
	}
}

func (m *monoOsc) calc(
	events [][]*midiEvent,
	oscParams *oscParams,
	adsrParams *adsrParams,
	lfoParams []*lfoParams,
	envelopeParams []*envelopeParams,
	glideTime int,
) {
	for i, lfo := range m.lfos {
		lfo.applyParams(lfoParams[i])
	}
	for i, envelope := range m.envelopes {
		envelope.destination = envelopeParams[i].destination
		envelope.adsr.applyEnvelopeParams(envelopeParams[i])
	}
	m.adsr.setParams(adsrParams)
	for i := int64(0); i < int64(len(m.out)); i++ {
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
						m.osc.initWithNote(oscParams, data.note)
					} else {
						m.osc.glide(oscParams, m.activeNotes[0], glideTime)
					}
					m.adsr.noteOn()
					for _, envelope := range m.envelopes {
						envelope.noteOn()
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
					m.osc.glide(oscParams, m.activeNotes[0], glideTime)
				} else {
					m.adsr.noteOff()
				}
				for _, envelope := range m.envelopes {
					envelope.noteOff()
				}
			}
		}
		for _, envelope := range m.envelopes {
			envelope.step()
		}
		m.adsr.step()
		freqRatio := 1.0
		phaseShift := 0.0
		ampRatio := 1.0
		for lfoIndex, lfo := range m.lfos {
			amountGain := 1.0
			lfoFreqRatio := 1.0
			for _, envelope := range m.envelopes {
				if envelope.destination == "lfo"+fmt.Sprint(lfoIndex)+"_amount" {
					amountGain *= envelope.value
				}
				if envelope.destination == "lfo"+fmt.Sprint(lfoIndex)+"_freq" {
					lfoFreqRatio *= math.Pow(16.0, envelope.value)
				}
			}
			_freqRatio, _phaseShift, _ampRatio := lfo.step(m.osc, amountGain, lfoFreqRatio)
			freqRatio *= _freqRatio
			phaseShift += _phaseShift
			ampRatio *= _ampRatio
		}
		for _, envelope := range m.envelopes {
			if envelope.destination == "freq" {
				freqRatio *= math.Pow(16.0, envelope.value)
			}
		}
		m.out[i] = m.osc.step(freqRatio, phaseShift) * oscGain * ampRatio * m.adsr.value
	}
}

// ----- POLY OSC ----- //

type polyOsc struct {
	// pooled + active = maxPoly
	pooled []*noteOsc
	active []*noteOsc
	out    []float64 // length: samplesPerCycle
}

func newPolyOsc() *polyOsc {
	pooled := make([]*noteOsc, maxPoly)
	for i := 0; i < len(pooled); i++ {
		pooled[i] = &noteOsc{
			osc:       &osc{phase01: rand.Float64()},
			adsr:      &adsr{},
			lfos:      []*lfo{newLfo(), newLfo(), newLfo()},
			envelopes: []*envelope{newEnvelope(), newEnvelope(), newEnvelope()},
		}
	}
	return &polyOsc{
		pooled: pooled,
		out:    make([]float64, samplesPerCycle),
	}
}

func (p *polyOsc) calc(
	events [][]*midiEvent,
	oscParams *oscParams,
	adsrParams *adsrParams,
	lfoParams []*lfoParams,
	envelopeParams []*envelopeParams,
) {
	for _, o := range p.active {
		for i, lfo := range o.lfos {
			lfo.applyParams(lfoParams[i])
		}
		for i, envelope := range o.envelopes {
			envelope.destination = envelopeParams[i].destination
			envelope.adsr.applyEnvelopeParams(envelopeParams[i])
		}
	}
	for _, o := range p.pooled {
		for i, lfo := range o.lfos {
			lfo.applyParams(lfoParams[i])
		}
		for i, envelope := range o.envelopes {
			envelope.destination = envelopeParams[i].destination
			envelope.adsr.applyEnvelopeParams(envelopeParams[i])
		}
	}
	for i := int64(0); i < int64(len(p.out)); i++ {
		p.out[i] = 0
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
					o.osc.initWithNote(oscParams, data.note)
					o.adsr.init(adsrParams)
				} else {
					log.Println("maxPoly exceeded")
				}
			}
		}

		for j := len(p.active) - 1; j >= 0; j-- {
			o := p.active[j]
			for _, e := range events {
				switch data := e.event.(type) {
				case *noteOff:
					if data.note == o.note {
						o.adsr.noteOff()
						for _, envelope := range o.envelopes {
							envelope.noteOff()
						}
					}
				case *noteOn:
					if data.note == o.note {
						o.adsr.noteOn()
						for _, envelope := range o.envelopes {
							envelope.noteOn()
						}
					}
				}
			}
			o.adsr.step()
			for _, envelope := range o.envelopes {
				envelope.step()
			}
			freqRatio := 1.0
			phaseShift := 0.0
			ampRatio := 1.0
			for lfoIndex, lfo := range o.lfos {
				amountGain := 1.0
				lfoFreqRatio := 1.0
				for _, envelope := range o.envelopes {
					if envelope.destination == "lfo"+fmt.Sprint(lfoIndex)+"_amount" {
						amountGain *= envelope.value
					}
					if envelope.destination == "lfo"+fmt.Sprint(lfoIndex)+"_freq" {
						lfoFreqRatio *= math.Pow(16.0, envelope.value)
					}
				}
				_freqRatio, _phaseShift, _ampRatio := lfo.step(o.osc, amountGain, lfoFreqRatio)
				freqRatio *= _freqRatio
				phaseShift += _phaseShift
				ampRatio *= _ampRatio
			}
			for _, envelope := range o.envelopes {
				if envelope.destination == "freq" {
					freqRatio *= math.Pow(16.0, envelope.value)
				}
			}
			p.out[i] += o.osc.step(freqRatio, phaseShift) * oscGain * ampRatio * o.adsr.value
			if o.adsr.phase == "" {
				p.active = append(p.active[:j], p.active[j+1:]...)
				p.pooled = append(p.pooled, o)
			}
		}
	}
}

// ----- POLY OSC V2 ----- //

type polyOsc2 struct {
	oscs []*noteOsc
	out  []float64 // length: samplesPerCycle
}

func newPolyOsc2() *polyOsc2 {
	oscs := make([]*noteOsc, 128)
	for i := 0; i < len(oscs); i++ {
		oscs[i] = &noteOsc{
			osc:       &osc{phase01: rand.Float64()},
			adsr:      &adsr{},
			lfos:      []*lfo{newLfo(), newLfo(), newLfo()},
			envelopes: []*envelope{newEnvelope(), newEnvelope(), newEnvelope()},
		}
	}
	return &polyOsc2{
		oscs: oscs,
		out:  make([]float64, samplesPerCycle),
	}
}

func (p *polyOsc2) calc(
	events [][]*midiEvent,
	oscParams *oscParams,
	adsrParams *adsrParams,
	lfoParams []*lfoParams,
	envelopeParams []*envelopeParams,
) {
	for _, o := range p.oscs {
		for i, lfo := range o.lfos {
			lfo.applyParams(lfoParams[i])
		}
		for i, envelope := range o.envelopes {
			envelope.destination = envelopeParams[i].destination
			envelope.adsr.applyEnvelopeParams(envelopeParams[i])
		}
	}
	for i := int64(0); i < int64(len(p.out)); i++ {
		p.out[i] = 0
		events := events[i]
		for j := 0; j < len(events); j++ {
			switch data := events[j].event.(type) {
			case *noteOn:
				o := p.oscs[data.note]
				o.note = data.note
				o.osc.initWithNoteWithoutRandomizePhase(oscParams, data.note)
				o.adsr.setParams(adsrParams)
				o.adsr.noteOn()
				for _, envelope := range o.envelopes {
					envelope.noteOn()
				}
			case *noteOff:
				o := p.oscs[data.note]
				o.adsr.noteOff()
				for _, envelope := range o.envelopes {
					envelope.noteOff()
				}
			}
		}
		for j := len(p.oscs) - 1; j >= 0; j-- {
			o := p.oscs[j]
			if o.adsr.phase == "" {
				continue
			}
			o.adsr.step()
			for _, envelope := range o.envelopes {
				envelope.step()
			}
			freqRatio := 1.0
			phaseShift := 0.0
			ampRatio := 1.0
			for lfoIndex, lfo := range o.lfos {
				amountGain := 1.0
				lfoFreqRatio := 1.0
				for _, envelope := range o.envelopes {
					if envelope.destination == "lfo"+fmt.Sprint(lfoIndex)+"_amount" {
						amountGain *= envelope.value
					}
					if envelope.destination == "lfo"+fmt.Sprint(lfoIndex)+"_freq" {
						lfoFreqRatio *= math.Pow(16.0, envelope.value)
					}
				}
				_freqRatio, _phaseShift, _ampRatio := lfo.step(o.osc, amountGain, lfoFreqRatio)
				freqRatio *= _freqRatio
				phaseShift += _phaseShift
				ampRatio *= _ampRatio
			}
			for _, envelope := range o.envelopes {
				if envelope.destination == "freq" {
					freqRatio *= math.Pow(16.0, envelope.value)
				}
			}
			p.out[i] += o.osc.step(freqRatio, phaseShift) * oscGain * ampRatio * o.adsr.value
		}
	}
}

// ----- NOTE OSC ----- //

type noteOsc struct {
	note      int
	osc       *osc
	adsr      *adsr
	lfos      []*lfo
	envelopes []*envelope
}

// ----- OSC ----- //

type oscParams struct {
	kind   string
	octave int // -2 ~ 2
	coarse int // -12 ~ 12
	fine   int // -100 ~ 100 cent
}
type oscJSON struct {
	Kind   string `json:"kind"`
	Octave int    `json:"octave"`
	Coarse int    `json:"coarse"`
	Fine   int    `json:"fine"`
}

func (o *oscParams) applyJSON(data json.RawMessage) {
	var j oscJSON
	err := json.Unmarshal(data, &j)
	if err != nil {
		log.Println("failed to apply JSON to oscParams")
		return
	}
	o.kind = j.Kind
	o.octave = j.Octave
	o.coarse = j.Coarse
	o.fine = j.Fine
}
func (o *oscParams) toJSON() json.RawMessage {
	return toRawMessage(&oscJSON{
		Kind:   o.kind,
		Octave: o.octave,
		Coarse: o.coarse,
		Fine:   o.fine,
	})
}
func (o *oscParams) set(key string, value string) error {
	switch key {
	case "kind":
		o.kind = value
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
	}
	return nil
}

type osc struct {
	kind      string
	glideTime int // ms
	freq      float64
	phase01   float64
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
	return noteToFreq(note) * math.Pow(2, float64(p.octave+p.coarse+p.fine/100/12))
}
func (o *osc) initWithNote(p *oscParams, note int) {
	o.kind = p.kind
	o.freq = noteWithParamsToFreq(p, note)
	o.phase01 = rand.Float64()
}
func (o *osc) initWithNoteWithoutRandomizePhase(p *oscParams, note int) {
	o.kind = p.kind
	o.freq = noteWithParamsToFreq(p, note)
}
func (o *osc) glide(p *oscParams, note int, glideTime int) {
	nextFreq := noteWithParamsToFreq(p, note)
	if math.Abs(nextFreq-o.freq) < 0.001 {
		return
	}
	o.glideTime = glideTime
	o.prevFreq = o.freq
	o.nextFreq = nextFreq
	o.gliding = true
	o.shiftPos = 0
}
func (o *osc) step(freqRatio float64, phaseShift float64) float64 {
	freq := o.freq * freqRatio
	p := positiveMod(o.phase01+phaseShift/(2.0*math.Pi), 1)
	value := 0.0
	switch o.kind {
	case "sine":
		value = math.Sin(2 * math.Pi * p)
	case "triangle":
		if p < 0.5 {
			value = p*4 - 1
		} else {
			value = p*(-4) + 3
		}
	case "square":
		if p < 0.5 {
			value = 1
		} else {
			value = -1
		}
	case "square-wt":
		note := freqToNote(freq)
		value = blsquareWT.tables[note].getAtPhase(2.0 * math.Pi * p)
	case "pulse":
		if p < 0.25 {
			value = 1
		} else {
			value = -1
		}
	case "saw":
		value = p*2 - 1
	case "saw-wt":
		note := freqToNote(freq)
		value = blsawWT.tables[note].getAtPhase(2.0 * math.Pi * p)
	case "saw-rev":
		value = p*(-2) + 1
	case "noise":
		value = rand.Float64()*2 - 1
	}
	o.phase01 += freq / float64(sampleRate)
	_, o.phase01 = math.Modf(o.phase01)
	if o.gliding {
		o.shiftPos++
		t := o.shiftPos * secPerSample * 1000 / float64(o.glideTime)
		o.freq = t*o.nextFreq + (1-t)*o.prevFreq
		if t >= 1 || math.Abs(o.nextFreq-o.freq) < 0.001 {
			o.freq = o.nextFreq
			o.gliding = false
		}
	}
	return value
}

// ----- ADSR ----- //

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
	phase          string // "attack", "keep", "decay", "sustain", "release" nil
	phasePos       int
	valueAtNoteOn  float64
	valueAtNoteOff float64
}

func (a *adsr) init(p *adsrParams) {
	a.setParams(p)
	a.value = 0
	a.phase = ""
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
	a.base = 1
	a.peek = 0
	if p.destination == "freq" ||
		p.destination == "filter_freq" ||
		p.destination == "lfo0_freq" ||
		p.destination == "lfo1_freq" ||
		p.destination == "lfo2_freq" {
		a.base = 0
		a.peek = p.amount
	}
	a.attack = 0
	a.keep = p.delay
	a.decay = p.attack
	a.sustain = a.base
	a.release = 0
}

func (a *adsr) noteOn() {
	a.phase = "attack"
	a.phasePos = 0
	a.valueAtNoteOn = a.value
}

func (a *adsr) noteOff() {
	a.phase = "release"
	a.phasePos = 0
	a.valueAtNoteOff = a.value
}

func (a *adsr) step() {
	phaseTime := float64(a.phasePos) * secPerSample * 1000 // ms
	switch a.phase {
	case "attack":
		if phaseTime >= float64(a.attack) {
			a.phase = "keep"
			a.phasePos = 0
			a.value = a.peek
		} else {
			t := phaseTime / float64(a.attack)
			a.value = t*a.peek + (1-t)*a.valueAtNoteOn // TODO: don't use the same attack time
			a.phasePos++
		}
	case "keep":
		if phaseTime >= float64(a.keep) {
			a.phase = "decay"
			a.phasePos = 0
		} else {
			a.phasePos++
		}
	case "decay":
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
			a.phase = "sustain"
			a.phasePos = 0
			a.value = float64(a.sustain)
		} else {
			a.phasePos++
		}
	case "sustain":
		a.value = float64(a.sustain)
	case "release":
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
			a.phase = ""
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

// ----- Filter ----- //

type filter struct {
	kind string
	freq float64
	q    float64
	gain float64
	N    int
	a    []float64 // feedforward
	b    []float64 // feedback
	past []float64
}
type filterJSON struct {
	Kind string  `json:"kind"`
	Freq float64 `json:"freq"`
	Q    float64 `json:"q"`
	Gain float64 `json:"gain"`
}

func (f *filter) applyJSON(data json.RawMessage) {
	var j filterJSON
	err := json.Unmarshal(data, &j)
	if err != nil {
		log.Println("failed to apply JSON to filter")
		return
	}
	f.kind = j.Kind
	f.freq = j.Freq
	f.q = j.Q
	f.gain = j.Gain
}
func (f *filter) toJSON() json.RawMessage {
	return toRawMessage(&filterJSON{
		Kind: f.kind,
		Freq: f.freq,
		Q:    f.q,
		Gain: f.gain,
	})
}

func (f *filter) process(in []float64, out []float64) {
	if len(f.a) == 0 {
		f.a, f.b = getH(f)
		f.past = make([]float64, int(math.Max(float64(len(f.a)-1), float64(len(f.b)))))
	}
	processFilter(in, out, f.a, f.b, f.past)
}

func getH(f *filter) ([]float64, []float64) {
	fc := f.freq / sampleRate
	switch f.kind {
	case "lowpass-fir":
		return makeFIRLowpassH(f.N, fc, hamming)
	case "highpass-fir":
		return makeFIRHighpassH(f.N, fc, hamming)
	case "lowpass":
		return makeBiquadLowpassH(fc, f.q)
	case "highpass":
		return makeBiquadHighpassH(fc, f.q)
	case "bandpass-1":
		return makeBiquadBandpass1H(fc, f.q)
	case "bandpass-2":
		return makeBiquadBandpass2H(fc, f.q)
	case "notch":
		return makeBiquadNotchH(fc, f.q)
	case "peaking":
		return makeBiquadPeakingEQH(fc, f.q, f.gain)
	case "lowshelf":
		return makeBiquadLowShelfH(fc, f.q, f.gain)
	case "highshelf":
		return makeBiquadHighShelfH(fc, f.q, f.gain)
	case "none":
		fallthrough
	default:
		return makeNoFilterH()
	}
}

func processFilter(in []float64, out []float64, a []float64, b []float64, past []float64) {
	for i := 0; i < len(in); i++ {
		out[i] = processFilterEach(in[i], a, b, past)
	}
}
func processFilterEach(in float64, a []float64, b []float64, past []float64) float64 {
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
	processFilter(in, out, a, b, past)
	return out
}

func frequencyResponse(a []float64, b []float64) []float64 {
	h := impulseResponse(a, b, fftSize)
	fft.CalcAbs(h)
	return h[:fftSize/2]
}

func (f *filter) set(key string, value string) error {
	switch key {
	case "kind":
		f.kind = value
		f.past = nil
		f.a = nil
		f.b = nil
	case "freq":
		freq, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		f.freq = freq
		f.past = nil
		f.a = nil
		f.b = nil
	case "q":
		q, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		f.q = q
		f.past = nil
		f.a = nil
		f.b = nil
	case "gain":
		gain, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		f.gain = gain
		f.past = nil
		f.a = nil
		f.b = nil
	}
	return nil
}

// ----- Envelope ----- //

type envelopeParams struct {
	destination string
	delay       float64
	attack      float64
	amount      float64
}

type envelopeJSON struct {
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
	l.destination = j.Destination
	l.delay = j.Delay
	l.attack = j.Attack
	l.amount = j.Amount
}
func (l *envelopeParams) toJSON() json.RawMessage {
	return toRawMessage(&envelopeJSON{
		Destination: l.destination,
		Delay:       l.delay,
		Attack:      l.attack,
		Amount:      l.amount,
	})
}
func newEnvelopeParams() *envelopeParams {
	return &envelopeParams{
		destination: "none",
		delay:       0,
		attack:      0,
		amount:      0,
	}
}
func (l *envelopeParams) set(key string, value string) error {
	switch key {
	case "destination":
		l.destination = value
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

type envelope struct {
	*adsr
	destination string
}

func newEnvelope() *envelope {
	return &envelope{
		destination: "none",
		adsr:        &adsr{},
	}
}

// ----- LFO ----- //

type lfoParams struct {
	destination string
	wave        string
	freqType    string
	freq        float64
	amount      float64
}

type lfoJSON struct {
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
	l.destination = j.Destination
	l.wave = j.Wave
	l.freqType = j.FreqType
	l.freq = j.Freq
	l.amount = j.Amount
}
func (l *lfoParams) toJSON() json.RawMessage {
	return toRawMessage(&lfoJSON{
		Destination: l.destination,
		Wave:        l.wave,
		FreqType:    l.freqType,
		Freq:        l.freq,
		Amount:      l.amount,
	})
}

func newLfoParams() *lfoParams {
	return &lfoParams{
		destination: "none",
		wave:        "sine",
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
	case "destination":
		// l.initByDestination(value)
		l.destination = value
	case "wave":
		l.wave = value
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
	destination string
	freqType    string
	amount      float64
	osc         *osc
}

func newLfo() *lfo {
	return &lfo{
		destination: "none",
		freqType:    "none",
		amount:      0,
		osc:         &osc{phase01: rand.Float64()},
	}
}

func (l *lfo) applyParams(p *lfoParams) {
	l.destination = p.destination
	l.osc.kind = p.wave
	l.freqType = p.freqType
	l.osc.freq = p.freq
	l.amount = p.amount
}

func (l *lfo) step(career *osc, amountGain float64, lfoFreqRatio float64) (float64, float64, float64) {
	freqRatio := 1.0
	phaseShift := 0.0
	ampRatio := 1.0
	switch l.destination {
	case "vibrato":
		amount := l.amount * amountGain
		freqRatio = math.Pow(2.0, l.osc.step(lfoFreqRatio, 0.0)*amount/100.0/12.0)
	case "tremolo":
		amount := l.amount * amountGain
		ampRatio = 1.0 + (l.osc.step(lfoFreqRatio, 0.0)-1.0)/2.0*amount
	case "fm":
		amount := l.amount * amountGain
		freqRatio = math.Pow(2.0, l.osc.step(career.freq*lfoFreqRatio, 0.0)*amount/100/12)
	case "pm":
		amount := l.amount * amountGain
		phaseShift = l.osc.step(career.freq*lfoFreqRatio, 0.0) * amount
	case "am":
		amount := l.amount * amountGain
		ampRatio = 1.0 + l.osc.step(career.freq*lfoFreqRatio, 0.0)*amount
	}
	return freqRatio, phaseShift, ampRatio
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
	oscParams      *oscParams
	adsrParams     *adsrParams
	lfoParams      []*lfoParams
	envelopeParams []*envelopeParams
	polyMode       bool
	glideTime      int // ms
	monoOsc        *monoOsc
	polyOsc        *polyOsc
	filter         *filter
	pos            int64
	out            []float64 // length: fftSize
	lastRead       float64
}
type stateJSON struct {
	Poly      string            `json:"poly"`
	GlideTime int               `json:"glideTime"`
	Osc       json.RawMessage   `json:"osc"`
	Adsr      json.RawMessage   `json:"adsr"`
	Lfos      []json.RawMessage `json:"lfos"`
	Envelopes []json.RawMessage `json:"envelopes"`
	Filter    json.RawMessage   `json:"filter"`
}

func (s *state) applyJSON(data json.RawMessage) {
	var j stateJSON
	err := json.Unmarshal(data, &j)
	if err != nil {
		log.Println("failed to apply JSON to state")
		return
	}
	s.polyMode = j.Poly == "poly"
	s.glideTime = j.GlideTime
	s.oscParams.applyJSON(j.Osc)
	s.adsrParams.applyJSON(j.Adsr)
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
	s.filter.applyJSON(j.Filter)
}
func (s *state) toJSON() json.RawMessage {
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
		Osc:       s.oscParams.toJSON(),
		Adsr:      s.adsrParams.toJSON(),
		Lfos:      lfoJsons,
		Envelopes: envelopeJsons,
		Filter:    s.filter.toJSON(),
	})
}

func newState() *state {
	return &state{
		events:         make([][]*midiEvent, samplesPerCycle*2),
		oscParams:      &oscParams{kind: "sine"},
		adsrParams:     &adsrParams{attack: 10, decay: 100, sustain: 0.7, release: 200},
		lfoParams:      []*lfoParams{newLfoParams(), newLfoParams(), newLfoParams()},
		envelopeParams: []*envelopeParams{newEnvelopeParams(), newEnvelopeParams(), newEnvelopeParams()},
		polyMode:       false,
		glideTime:      100,
		monoOsc:        newMonoOsc(),
		polyOsc:        newPolyOsc(),
		filter:         &filter{kind: "none", freq: 1000, q: 1, gain: 0, N: 50},
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
		if a.state.polyMode {
			a.state.polyOsc.calc(a.state.events, a.state.oscParams, a.state.adsrParams, a.state.lfoParams, a.state.envelopeParams)
			a.state.filter.process(a.state.polyOsc.out, out)
		} else {
			a.state.monoOsc.calc(a.state.events, a.state.oscParams, a.state.adsrParams, a.state.lfoParams, a.state.envelopeParams, a.state.glideTime)
			a.state.filter.process(a.state.monoOsc.out, out)
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
		audio.update(command)
	}
	log.Println("processCommands() ended.")
}

func (a *Audio) update(command []string) {
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
				panic(err)
			}
			a.state.glideTime = int(value)
		case "osc":
			command = command[1:]
			if len(command) != 2 {
				panic(fmt.Errorf("invalid key-value pair %v", command))
			}
			err := a.state.oscParams.set(command[0], command[1])
			if err != nil {
				panic(err)
			}
		case "adsr":
			command = command[1:]
			if len(command) != 2 {
				panic(fmt.Errorf("invalid key-value pair %v", command))
			}
			err := a.state.adsrParams.set(command[0], command[1])
			if err != nil {
				panic(err)
			}
		case "filter":
			command = command[1:]
			if len(command) != 2 {
				panic(fmt.Errorf("invalid key-value pair %v", command))
			}
			err := a.state.filter.set(command[0], command[1])
			if err != nil {
				panic(err)
			}
			a.Changes.Add("filter-shape")
		case "lfo":
			command = command[1:]
			index, err := strconv.ParseInt(command[0], 10, 64)
			if err != nil {
				panic(err)
			}
			command = command[1:]
			if len(command) != 2 {
				panic(fmt.Errorf("invalid key-value pair %v", command))
			}
			err = a.state.lfoParams[index].set(command[0], command[1])
			if err != nil {
				panic(err)
			}
		case "envelope":
			command = command[1:]
			index, err := strconv.ParseInt(command[0], 10, 64)
			if err != nil {
				panic(err)
			}
			command = command[1:]
			if len(command) != 2 {
				panic(fmt.Errorf("invalid key-value pair %v", command))
			}
			err = a.state.envelopeParams[index].set(command[0], command[1])
			if err != nil {
				panic(err)
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
			panic(err)
		}
		a.addMidiEvent(&noteOn{note: int(note)})
	case "note_off":
		note, err := strconv.ParseInt(command[1], 10, 32)
		if err != nil {
			panic(err)
		}
		a.addMidiEvent(&noteOff{note: int(note)})
	default:
		panic(fmt.Errorf("unknown command %v", command[0]))
	}
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

// GetFilterShape ...
func (a *Audio) GetFilterShape() []float64 {
	a.state.Lock()
	_a, b := getH(a.state.filter)
	a.state.Unlock()
	return frequencyResponse(_a, b)
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
