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

// ----- OSC ----- //

type osc struct {
	kind string
	freq float64
	out  []float64 // length: samplesPerCycle
}

func (o *osc) calc(pos int64, events [][]*midiEvent) {
	for i := int64(0); i < int64(len(o.out)); i++ {
		if events := events[i]; events != nil {
			for j := 0; j < len(events); j++ {
				switch data := events[j].event.(type) {
				case *noteOn:
					o.freq = 442 * math.Pow(2, float64(data.note-69)/12)
				}
			}
		}
		o.out[i] = o.calcEach(pos+i) * 0.3
	}
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
	case "pluse":
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
func (o *osc) set(key string, value string) error {
	switch key {
	case "kind":
		o.kind = value
	case "freq":
		freq, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		o.freq = freq
	}
	return nil
}

// ----- ADSR ----- //

type adsr struct {
	attack  float64
	decay   float64
	sustain float64
	release float64
	on      bool
}

func (a *adsr) calc(pos int64, events [][]*midiEvent, out []float64) {
	for i := int64(0); i < int64(len(out)); i++ {
		if events := events[i]; events != nil {
			for j := 0; j < len(events); j++ {
				switch events[j].event.(type) {
				case *noteOff:
					a.on = false
				case *noteOn:
					a.on = true
				}
			}
		}
		if a.on {
			out[i] *= 1
		} else {
			out[i] *= 0
		}
	}
}

func (a *adsr) set(key string, value string) error {
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

func (f *filter) Process(in []float64, out []float64) {
	if len(f.a) == 0 {
		f.a, f.b = getH(f)
		f.past = make([]float64, int(math.Max(float64(len(f.a)-1), float64(len(f.b)))))
	}
	process(in, out, f.a, f.b, f.past)
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

func process(in []float64, out []float64, a []float64, b []float64, past []float64) {
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
	process(in, out, a, b, past)
	return out
}

func frequencyResponse(a []float64, b []float64) []float64 {
	h := impulseResponse(a, b, fftSize)
	fft.CalcAbs(h)
	return h[:fftSize/2]
}

func (f *filter) Set(key string, value string) error {
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
	events   [][]*midiEvent // length: samplesPerCycle * 2
	osc      *osc
	adsr     *adsr
	filter   *filter
	pos      int64
	out      []float64 // length: fftSize
	lastRead float64
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
		a.state.osc.calc(a.state.pos, a.state.events)
		a.state.adsr.calc(a.state.pos, a.state.events, a.state.osc.out)
		offset := a.state.pos % fftSize
		out := a.state.out[offset : offset+bufSamples]
		a.state.filter.Process(a.state.osc.out, out)
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
			events: make([][]*midiEvent, samplesPerCycle*2),
			osc:    &osc{kind: "sine", freq: 442, out: make([]float64, samplesPerCycle)},
			adsr:   &adsr{},
			filter: &filter{kind: "none", freq: 1000, q: 1, gain: 0, N: 50},
			pos:    0,
			out:    make([]float64, fftSize),
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
			a.state.osc.set(command[0], command[1])
		case "adsr":
			command = command[1:]
			if len(command) != 2 {
				panic(fmt.Errorf("invalid key-value pair %v", command))
			}
			a.state.adsr.set(command[0], command[1])
		case "filter":
			command = command[1:]
			if len(command) != 2 {
				panic(fmt.Errorf("invalid key-value pair %v", command))
			}
			a.state.filter.Set(command[0], command[1])
			a.Changes.Add("filter-shape")
		}
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
