package audio

import (
	"context"
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

var fft = NewFFT(fftSize, false)

// ----- Utility ----- //

func now() float64 {
	return float64(time.Now().UnixNano()) / 1000 / 1000 / 1000
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
	osc  *osc
	adsr *adsr
	out  []float64 // length: samplesPerCycle
}

func newMonoOsc() *monoOsc {
	return &monoOsc{
		osc:  &osc{},
		adsr: &adsr{},
		out:  make([]float64, samplesPerCycle),
	}
}

func (m *monoOsc) calc(pos int64, events [][]*midiEvent, oscParams *oscParams, adsrParams *adsrParams) {
	m.adsr.setParams(adsrParams)
	for i := int64(0); i < int64(len(m.out)); i++ {
		m.out[i] = 0
		for _, e := range events[i] {
			switch data := e.event.(type) {
			case *noteOn:
				m.osc.initWithNote(oscParams, data.note)
				m.adsr.noteOn()
			case *noteOff:
				m.adsr.noteOff()
			}
		}
		m.adsr.step()
		m.out[i] += m.osc.calcEach(pos+i) * 0.1 * m.adsr.value
	}
}

// ----- POLY OSC ----- //

type polyOsc struct {
	// pooled + active = maxPoly
	pooled []*oscWithADSR
	active []*oscWithADSR
	out    []float64 // length: samplesPerCycle
}

func newPolyOsc() *polyOsc {
	pooled := make([]*oscWithADSR, maxPoly)
	for i := 0; i < len(pooled); i++ {
		pooled[i] = &oscWithADSR{
			osc:  &osc{},
			adsr: &adsr{},
		}
	}
	return &polyOsc{
		pooled: pooled,
		out:    make([]float64, samplesPerCycle),
	}
}

func (p *polyOsc) calc(pos int64, events [][]*midiEvent, oscParams *oscParams, adsrParams *adsrParams) {
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
			value := o.osc.calcEach(pos+int64(i)) * 0.1
			for _, e := range events {
				switch data := e.event.(type) {
				case *noteOff:
					if data.note == o.note {
						o.adsr.noteOff()
					}
				case *noteOn:
					if data.note == o.note {
						o.adsr.noteOn()
					}
				}
			}
			o.adsr.step()
			p.out[i] += value * o.adsr.value
			if o.adsr.phase == "" {
				p.active = append(p.active[:j], p.active[j+1:]...)
				p.pooled = append(p.pooled, o)
			}
		}
	}
}

// ----- OSC WITH ADSR ----- //

type oscWithADSR struct {
	note int
	osc  *osc
	adsr *adsr
}

// ----- OSC ----- //

type oscParams struct {
	kind   string
	octave int // -2 ~ 2
	coarse int // -12 ~ 12
	fine   int // -100 ~ 100 cent
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
	kind string
	freq float64
	out  []float64 // length: samplesPerCycle
}

func (o *osc) initWithNote(p *oscParams, note int) {
	o.kind = p.kind
	o.freq = 442 * math.Pow(2, float64(note+p.octave*12+p.coarse+p.fine/100-69)/12)
}

func (o *osc) calcEach(pos int64) float64 {
	switch o.kind {
	case "sine":
		length := float64(sampleRate) / float64(o.freq)
		return math.Sin(2 * math.Pi * float64(pos) / length)
	case "triangle":
		length := int64(float64(sampleRate) / float64(o.freq))
		if pos%length < length/2 {
			return float64(pos%length)/float64(length)*4 - 1
		}
		return float64(pos%length)/float64(length)*(-4) + 3
	case "square":
		length := int64(float64(sampleRate) / float64(o.freq))
		if pos%length < length/2 {
			return 1
		}
		return -1
	case "pulse":
		length := int64(float64(sampleRate) / float64(o.freq))
		if pos%length < length/4 {
			return 1
		}
		return -1
	case "saw":
		length := int64(float64(sampleRate) / float64(o.freq))
		return float64(pos%length)/float64(length)*2 - 1
	case "noise":
		return rand.Float64()*2 - 1
	}
	return 0
}

// ----- ADSR ----- //

type adsrParams struct {
	attack  float64 // ms
	decay   float64 // ms
	sustain float64 // 0-1
	release float64 // ms
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

type adsr struct {
	attack       float64 // ms
	decay        float64 // ms
	sustain      float64 // 0-1
	release      float64 // ms
	value        float64
	phase        string // "attack", "decay", "sustain", "release" nil
	phasePos     int
	releaseValue float64
}

func (a *adsr) init(p *adsrParams) {
	a.setParams(p)
	a.value = 0
	a.phase = ""
	a.phasePos = 0
	a.releaseValue = 0
}

func (a *adsr) setParams(p *adsrParams) {
	a.attack = p.attack
	a.decay = p.decay
	a.sustain = p.sustain
	a.release = p.release
}

func (a *adsr) calc(pos int64, events [][]*midiEvent, out []float64) {
	for i := range out {
		for _, e := range events[i] {
			switch e.event.(type) {
			case *noteOff:
				a.noteOff()
			case *noteOn:
				a.noteOn()
			}
		}
		a.step()
		out[i] *= a.value
	}
}

func (a *adsr) noteOn() {
	if a.phase == "" {
		a.phase = "attack"
		a.phasePos = 0
		a.value = 0
	}
}

func (a *adsr) noteOff() {
	a.phase = "release"
	a.phasePos = 0
	a.releaseValue = a.value
}

func (a *adsr) step() {
	t := float64(a.phasePos) * secPerSample * 1000 // ms
	switch a.phase {
	case "attack":
		a.value = t / float64(a.attack)
		if t >= float64(a.attack) {
			a.phase = "decay"
			a.phasePos = 0
			a.value = 1
		} else {
			a.phasePos++
		}
	case "decay":
		a.value = setTargetAtTime(1.0, a.sustain, t/float64(a.decay))
		if a.value-a.sustain < 0.001 {
			a.phase = "sustain"
			a.phasePos = 0
			a.value = float64(a.sustain)
		} else {
			a.phasePos++
		}
	case "sustain":
		a.value = float64(a.sustain)
	case "release":
		a.value = setTargetAtTime(a.releaseValue, 0.0, t/float64(a.release))
		if a.value < 0.001 {
			a.phase = ""
			a.phasePos = 0
			a.value = 0
		} else {
			a.phasePos++
		}
	default:
		a.value = 0
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
		// get input
		tmp := in[i]
		// apply b
		for j := 0; j < len(b); j++ {
			tmp -= past[j] * b[j]
		}
		// apply a
		o := tmp * a[0]
		for j := 1; j < len(a); j++ {
			o += past[j-1] * a[j]
		}
		// unshift f.past
		for j := len(past) - 2; j >= 0; j-- {
			past[j+1] = past[j]
		}
		if len(past) > 0 {
			past[0] = tmp
		}
		// set output
		out[i] = o
	}
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

type state struct {
	sync.Mutex
	events     [][]*midiEvent // length: samplesPerCycle * 2
	oscParams  *oscParams
	adsrParams *adsrParams
	polyMode   bool
	monoOsc    *monoOsc
	polyOsc    *polyOsc
	filter     *filter
	pos        int64
	out        []float64 // length: fftSize
	lastRead   float64
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
			a.state.polyOsc.calc(a.state.pos, a.state.events, a.state.oscParams, a.state.adsrParams)
			a.state.filter.process(a.state.polyOsc.out, out)
		} else {
			a.state.monoOsc.calc(a.state.pos, a.state.events, a.state.oscParams, a.state.adsrParams)
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
		state: &state{
			events:     make([][]*midiEvent, samplesPerCycle*2),
			oscParams:  &oscParams{kind: "sine"},
			adsrParams: &adsrParams{attack: 10, decay: 100, sustain: 0.7, release: 200},
			polyMode:   false,
			monoOsc:    newMonoOsc(),
			polyOsc:    newPolyOsc(),
			filter:     &filter{kind: "none", freq: 1000, q: 1, gain: 0, N: 50},
			pos:        0,
			out:        make([]float64, fftSize),
		},
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
		case "osc":
			command = command[1:]
			if len(command) != 2 {
				panic(fmt.Errorf("invalid key-value pair %v", command))
			}
			a.state.oscParams.set(command[0], command[1])
		case "adsr":
			command = command[1:]
			if len(command) != 2 {
				panic(fmt.Errorf("invalid key-value pair %v", command))
			}
			a.state.adsrParams.set(command[0], command[1])
		case "filter":
			command = command[1:]
			if len(command) != 2 {
				panic(fmt.Errorf("invalid key-value pair %v", command))
			}
			a.state.filter.set(command[0], command[1])
			a.Changes.Add("filter-shape")
		}
	case "mono":
		a.state.polyMode = false
	case "poly":
		a.state.polyMode = true
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
