package audio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
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
func velocityToGain(velocity int, velSense float64) float64 {
	return 1.0 - (1.0-float64(velocity)/127.0)*velSense
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
	note     int
	velocity int
}
type noteOff struct {
	note int
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
	*params
	events      [][]*midiEvent // length: samplesPerCycle * 2
	monoOsc     *monoOsc
	polyOsc     *polyOsc
	echo        *echo
	pos         int64
	out         []float64 // length: fftSize
	lastRead    float64
	processTime float64
}

func newState() *state {
	return &state{
		events:  make([][]*midiEvent, samplesPerCycle*2),
		params:  newParams(),
		monoOsc: newMonoOsc(),
		polyOsc: newPolyOsc(),
		echo:    &echo{delay: &delay{}},
		pos:     0,
		out:     make([]float64, fftSize),
	}
}

// ----- Audio ----- //

// Audio ...
type Audio struct {
	ctx           context.Context
	otoContext    *oto.Context
	presetManager *presetManager
	CommandCh     chan []string
	state         *state
	Changes       *Changes
	fftResult     []float64 // length: fftSize
}

var _ io.Reader = (*Audio)(nil)

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
			a.state.polyOsc.calc(a.state.events, a.state.oscParams, a.state.adsrParams, a.state.noteFilterParams, a.state.filterParams, a.state.formantParams, a.state.lfoParams, a.state.envelopeParams, a.state.velSense, a.state.echo, out)
		} else {
			a.state.monoOsc.calc(a.state.events, a.state.oscParams, a.state.adsrParams, a.state.noteFilterParams, a.state.filterParams, a.state.formantParams, a.state.lfoParams, a.state.envelopeParams, a.state.velSense, a.state.glideTime, a.state.echo, out)
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
func NewAudio(presetDir string) (*Audio, error) {
	otoContext, err := oto.NewContext(sampleRate, channelNum, bitDepthInBytes, bufferSizeInBytes)
	if err != nil {
		return nil, err
	}
	commandCh := make(chan []string, 256)
	audio := &Audio{
		ctx:           context.Background(),
		otoContext:    otoContext,
		presetManager: newPresetManager(presetDir),
		CommandCh:     commandCh,
		state:         newState(),
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
	switch command[0] {
	case "set":
		a.state.Lock()
		defer a.state.Unlock()
		command = command[1:]
		switch command[0] {
		case "glide_time":
			command = command[1:]
			value, err := strconv.ParseInt(command[0], 10, 64)
			if err != nil {
				return err
			}
			a.state.glideTime = int(value)
		case "vel_sense":
			command = command[1:]
			value, err := strconv.ParseFloat(command[0], 64)
			if err != nil {
				return err
			}
			a.state.velSense = value
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
		case "note_filter":
			command = command[1:]
			if len(command) != 2 {
				return fmt.Errorf("invalid key-value pair %v", command)
			}
			err := a.state.noteFilterParams.set(command[0], command[1])
			if err != nil {
				return err
			}
			a.Changes.Add("filter-shape")
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
		case "formant":
			command = command[1:]
			if len(command) != 2 {
				return fmt.Errorf("invalid key-value pair %v", command)
			}
			err := a.state.formantParams.set(command[0], command[1])
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
		a.state.Lock()
		defer a.state.Unlock()
		a.state.polyMode = false
		a.Changes.Add("data")
	case "poly":
		a.state.Lock()
		defer a.state.Unlock()
		a.state.polyMode = true
		a.Changes.Add("data")
	case "note_on":
		a.state.Lock()
		defer a.state.Unlock()
		note, err := strconv.ParseInt(command[1], 10, 32)
		if err != nil {
			return err
		}
		a.addMidiEvent(&noteOn{note: int(note), velocity: 127})
	case "note_off":
		a.state.Lock()
		defer a.state.Unlock()
		note, err := strconv.ParseInt(command[1], 10, 32)
		if err != nil {
			return err
		}
		a.addMidiEvent(&noteOff{note: int(note)})
	case "preset":
		command = command[1:]
		switch command[0] {
		case "list":
			a.Changes.Add("preset_list")
		case "load":
			a.Changes.Add("preset")
		case "save":
			a.Changes.Add("preset_list")
		case "save_as":
			a.Changes.Add("preset_list")
		case "delete":
			a.Changes.Add("preset_list")
		}
	default:
		return fmt.Errorf("unknown command %v", command[0])
	}
	return nil
}

// RestoreLastParams ...
func (a *Audio) RestoreLastParams() error {
	a.state.Lock() // TODO: too long lock
	defer a.state.Unlock()
	found, err := a.presetManager.restoreLastParams(a.state.params)
	if err != nil {
		return err
	}
	if found {
		log.Println("loaded temporary file in ", a.presetManager.dir)
	} else {
		log.Println("temporary file not found in ", a.presetManager.dir)
	}
	a.Changes.Add("all_params") // always exists
	return nil
}

// SaveTemporaryData ...
func (a *Audio) SaveTemporaryData() error {
	a.state.Lock() // TODO: too long lock
	defer a.state.Unlock()
	err := a.presetManager.saveTemporaryData(a.state.params)
	if err != nil {
		return err
	}
	log.Println("saved temporary file in ", a.presetManager.dir)
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
	filter := &filter{}
	filter.applyParams(a.state.filterParams)
	formant := newFormant()
	formant.applyParams(a.state.formantParams)
	a.state.Unlock()

	out := make([]float64, fftSize)
	for i := 0; i < fftSize; i++ {
		in := 0.0
		if i == 0 {
			in = 1.0
		}
		out[i] += formant.step(filter.step(in, 1.0, nil))
	}
	fft.CalcAbs(out)
	return out[:fftSize/2]
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
		velocity := int(data[2])
		a.addMidiEvent(&noteOn{note: note, velocity: velocity})
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
