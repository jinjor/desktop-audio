package audio

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"strconv"

	"github.com/hajimehoshi/oto"
)

const (
	sampleRate        = 48000
	channelNum        = 2
	bitDepthInBytes   = 2
	bufferSizeInBytes = 4096
	samplesPerCycle   = 128
	fftSize           = 2048
)
const bytesPerSample = bitDepthInBytes * channelNum
const byteLengthPerCycle = samplesPerCycle * bytesPerSample

var fft = NewFFT(fftSize, false)

// ----- OSC ----- //

type osc struct {
	kind string
	freq float64
}

func (o *osc) Calc(pos int64) float64 {
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
func (o *osc) Set(key string, value string) error {
	switch key {
	case "kind":
		o.kind = value
	case "freq":
		freq, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		o.SetFreq(freq)
	}
	return nil
}
func (o *osc) SetFreq(freq float64) {
	o.freq = freq
}

// ----- Filter ----- //

type filter struct {
	kind  string
	freq  float64
	h     []float64
	lastX []float64
	lastY []float64
}

func (f *filter) Process(in []float64, out []float64) {
	length := len(in)
	N := 10
	fc := f.freq / sampleRate
	switch f.kind {
	case "lowpass":
		if len(f.h) == 0 {
			f.h = make([]float64, N+1)
			calcFIRLowpass(N, fc, f.h)
		}
	case "highpass":
		if len(f.h) == 0 {
			f.h = make([]float64, N+1)
			calcFIRLowpass(N, fc, f.h)
			for i := 0; i < len(f.h); i++ {
				f.h[i] = sinc(math.Pi*float64(i)) - f.h[i]
			}
		}
	default:
		copy(out, in)
		return
	}
	for i := 0; i < length; i++ {
		out[i] = 0
		for j := 0; j <= N; j++ {
			var x float64
			if i-j >= 0 {
				x = in[i-j]
			} else {
				x = f.lastX[N+i-j]
			}
			out[i] += x * f.h[j]
		}
	}
	copy(f.lastX, in[length-N:])
	copy(f.lastY, out[length-N:])
}
func calcFIRLowpass(N int, fc float64, h []float64) {
	if N%2 != 0 {
		log.Panicf("N should be even")
	}
	if len(h) != N+1 {
		log.Panicf("length of h must be N + 1")
	}
	wc := 2 * math.Pi * fc
	for i := 0; i <= N; i++ {
		h[i] = 2 * fc * sinc(wc*float64(i-N/2))
	}
}
func sinc(x float64) float64 {
	if math.Abs(x) < 0.000000001 {
		return 1
	}
	return math.Sin(x) / x
}

func (f *filter) Set(key string, value string) error {
	switch key {
	case "kind":
		f.kind = value
		f.h = nil
	case "freq":
		freq, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		f.SetFreq(freq)
		f.h = nil
	}
	return nil
}
func (f *filter) SetFreq(freq float64) {
	f.freq = freq
}

// ----- Audio ----- //

// Audio ...
type Audio struct {
	otoContext *oto.Context
	CommandCh  chan []string
	ctx        context.Context
	stateCh    chan *state
}

var _ io.Reader = (*Audio)(nil)

type state struct {
	osc       *osc
	filter    *filter
	gain      float64
	pos       int64
	oscOut    []float64 // length: samplesPerCycle
	out       []float64 // length: fftSize
	fftResult []float64 // length: fftSize
}

func (a *Audio) Read(buf []byte) (int, error) {
	select {
	case <-a.ctx.Done():
		log.Println("Read() interrupted.")
		return 0, io.EOF
	case state := <-a.stateCh:
		defer func() { a.stateCh <- state }()
		sampleLength := int64(len(buf) / bytesPerSample)
		for i := int64(0); i < sampleLength; i++ {
			state.oscOut[i] = state.osc.Calc(state.pos+i) * state.gain
		}
		offset := state.pos % fftSize
		out := state.out[offset : offset+sampleLength]
		state.filter.Process(state.oscOut, out)
		writeBuffer(state.out, offset, buf, 0)
		writeBuffer(state.out, offset, buf, 1)
		state.pos += sampleLength
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
	stateCh := make(chan *state, 1)
	stateCh <- &state{
		osc:       &osc{kind: "sine", freq: 442},
		filter:    &filter{kind: "none", freq: 1000, lastX: make([]float64, 10), lastY: make([]float64, 10)},
		gain:      0,
		pos:       0,
		oscOut:    make([]float64, samplesPerCycle),
		out:       make([]float64, fftSize),
		fftResult: make([]float64, fftSize),
	}
	audio := &Audio{
		otoContext: otoContext,
		CommandCh:  commandCh,
		ctx:        context.Background(),
		stateCh:    stateCh,
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
	state := <-a.stateCh
	defer func() { a.stateCh <- state }()

	switch command[0] {
	case "set":
		command = command[1:]
		switch command[0] {
		case "osc":
			command = command[1:]
			if len(command) != 2 {
				panic(fmt.Errorf("invalid key-value pair %v", command))
			}
			state.osc.Set(command[0], command[1])
		case "filter":
			command = command[1:]
			if len(command) != 2 {
				panic(fmt.Errorf("invalid key-value pair %v", command))
			}
			state.filter.Set(command[0], command[1])
		}
	case "note_on":
		note, err := strconv.ParseInt(command[1], 10, 32)
		if err != nil {
			panic(err)
		}
		state.osc.SetFreq(442 * math.Pow(2, float64(note-69)/12))
		state.gain = 0.3
	case "note_off":
		state.gain = 0
	default:
		panic(fmt.Errorf("unknown command %v", command[0]))
	}
}

// Close ...
func (a *Audio) Close() error {
	log.Println("Closing Audio...")
	close(a.CommandCh)
	close(a.stateCh)
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
	if _, err := io.CopyBuffer(p, a, make([]byte, byteLengthPerCycle)); err != nil {
		return err
	}
	log.Println("Start() ended.")
	return nil
}

// GetFFT ...
func (a *Audio) GetFFT(ctx context.Context) []float64 {
	select {
	case <-ctx.Done():
		log.Println("GetFFT() interrupted.")
		return nil
	case state := <-a.stateCh:
		defer func() { a.stateCh <- state }()
		// out:       | 4 | 1 | 2 | 3 |
		// offset:        ^
		// fftResult: | 1 | 2 | 3 | 4 |
		// return:    |<----->|
		offset := state.pos % fftSize
		copy(state.fftResult, state.out[offset:])
		copy(state.fftResult[fftSize-offset:], state.out[:offset])
		Han(state.fftResult)
		fft.CalcAbs(state.fftResult)
		for i, value := range state.fftResult {
			state.fftResult[i] = value * 2 / fftSize
		}
		return state.fftResult[:fftSize/2]
	}
}
