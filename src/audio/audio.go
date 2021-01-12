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
)
const bytesPerSample = bitDepthInBytes * channelNum
const byteLengthPerCycle = samplesPerCycle * bytesPerSample

var fft = NewFFT(samplesPerCycle, false)

// ----- OSC ----- //

type osc struct {
	freq float64
	kind string
}

func (s *osc) Calc(pos int64) float64 {
	switch s.kind {
	case "sine":
		length := float64(sampleRate) / float64(s.freq)
		return math.Sin(2 * math.Pi * float64(pos) / length)
	case "triangle":
		length := int64(float64(sampleRate) / float64(s.freq))
		if pos%length < length/2 {
			return float64(pos%length)/float64(length)*4 - 1
		}
		return float64(pos%length)/float64(length)*(-4) + 3
	case "square":
		length := int64(float64(sampleRate) / float64(s.freq))
		if pos%length < length/2 {
			return 1
		}
		return -1
	case "pluse":
		length := int64(float64(sampleRate) / float64(s.freq))
		if pos%length < length/4 {
			return 1
		}
		return -1
	case "saw":
		length := int64(float64(sampleRate) / float64(s.freq))
		return float64(pos%length)/float64(length)*2 - 1
	case "noise":
		return rand.Float64()*2 - 1
	}
	return 0
}
func (s *osc) Set(key string, value string) error {
	switch key {
	case "freq":
		freq, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		s.SetFreq(freq)
	case "kind":
		s.kind = value
	}
	return nil
}
func (s *osc) SetFreq(freq float64) {
	s.freq = freq
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
	osc  *osc
	gain float64
	pos  int64
	out  []float64
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
			state.out[i] = state.osc.Calc(state.pos+i) * state.gain
		}
		writeBuffer(state.out, buf, 0)
		writeBuffer(state.out, buf, 1)
		state.pos += sampleLength
		return len(buf), nil // io.EOF, etc.
	}
}

func writeBuffer(out []float64, buf []byte, ch int) {
	sampleLength := int(len(buf) / bytesPerSample)
	for i := 0; i < sampleLength; i++ {
		switch bitDepthInBytes {
		case 1:
			const max = 127
			b := int(out[i] * max)
			buf[bytesPerSample*i+ch] = byte(b + 128)
		case 2:
			const max = 32767
			b := int16(out[i] * max)
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
		osc:  &osc{freq: 442, kind: "sine"},
		gain: 0,
		out:  make([]float64, samplesPerCycle),
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
			if len(command) != 2 {
				panic(fmt.Errorf("invalid key-value pair %v", command))
			}
			state.osc.Set(command[0], command[1])
		}
		state.osc.Set("kind", command[1])
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
		return fft.CalcAbs(state.out)
	}
}
