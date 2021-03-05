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
	filterParams   *filterParams
	formantParams  *formantParams
	lfoParams      []*lfoParams
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
	Filter    json.RawMessage   `json:"filter"`
	Formant   json.RawMessage   `json:"formant"`
	Lfos      []json.RawMessage `json:"lfos"`
	Envelopes []json.RawMessage `json:"envelopes"`
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
	s.formantParams.applyJSON(j.Formant)
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
		Filter:    s.filterParams.toJSON(),
		Formant:   s.formantParams.toJSON(),
		Lfos:      lfoJsons,
		Envelopes: envelopeJsons,
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
		formantParams:  &formantParams{kind: formantA, tone: 1, q: 1},
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
			a.state.polyOsc.calc(a.state.events, a.state.oscParams, a.state.adsrParams, a.state.filterParams, a.state.formantParams, a.state.lfoParams, a.state.envelopeParams, a.state.echo, out)
		} else {
			a.state.monoOsc.calc(a.state.events, a.state.oscParams, a.state.adsrParams, a.state.filterParams, a.state.formantParams, a.state.lfoParams, a.state.envelopeParams, a.state.glideTime, a.state.echo, out)
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
	// TODO: formant
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
