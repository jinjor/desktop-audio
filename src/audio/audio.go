package audio

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"time"

	"github.com/hajimehoshi/oto"
)

const (
	sampleRate         = 44100
	channelNum         = 2
	bitDepthInBytes    = 2
	bufferSizeInBytes  = 4096
	byteLengthPerCycle = 128
)

type osc interface {
	Calc(int64) float64
	Set(string, string) error
	SetFreq(float64)
}

type oscImpl struct {
	freq float64
	kind string
}

var _ osc = (*oscImpl)(nil)

func (s *oscImpl) Calc(pos int64) float64 {
	switch s.kind {
	case "sine":
		length := float64(sampleRate) / float64(s.freq)
		return math.Sin(2 * math.Pi * float64(pos) / length)
	case "square":
		length := int64(float64(sampleRate) / float64(s.freq))
		if pos%length > length/2 {
			return 1
		}
		return 0
	}
	return 0
}
func (s *oscImpl) Set(key string, value string) error {
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
func (s *oscImpl) SetFreq(freq float64) {
	s.freq = freq
}

type state struct {
	osc       osc
	gain      float64
	pos       int64
	remaining []byte
}

func newState() *state {
	return &state{
		osc:  &oscImpl{freq: 442, kind: "sine"},
		gain: 0,
	}
}

func (s *state) Read(buf []byte) (int, error) {
	calc := func(p int64) float64 {
		return s.osc.Calc(p) * s.gain
	}
	writeBuffer(buf, s.pos, channelNum, bitDepthInBytes, calc)
	s.pos += int64(len(buf))
	return len(buf), nil // io.EOF, etc.
}

func writeBuffer(buf []byte, start int64, channelNum int, bitDepthInBytes int, calc func(int64) float64) {
	num := bitDepthInBytes * channelNum
	p := start / int64(num)
	switch bitDepthInBytes {
	case 1:
		for i := 0; i < len(buf)/num; i++ {
			const max = 127
			b := int(calc(p) * max)
			for ch := 0; ch < channelNum; ch++ {
				buf[num*i+ch] = byte(b + 128)
			}
			p++
		}
	case 2:
		for i := 0; i < len(buf)/num; i++ {
			const max = 32767
			b := int16(calc(p) * max)
			for ch := 0; ch < channelNum; ch++ {
				buf[num*i+2*ch] = byte(b)
				buf[num*i+1+2*ch] = byte(b >> 8)
			}
			p++
		}
	}
}

// Loop ...
func Loop(ch chan []string) {
	s := newState()
	c, err := oto.NewContext(sampleRate, channelNum, bitDepthInBytes, bufferSizeInBytes)
	if err != nil {
		panic(err)
	}
	go func() {
		p := c.NewPlayer()
		if _, err := io.CopyBuffer(p, s, make([]byte, byteLengthPerCycle)); err != nil {
			panic(err)
		}
		if err := p.Close(); err != nil {
			panic(err)
		}
	}()
	for range time.Tick(1 * time.Millisecond) {
		select {
		case command := <-ch:
			switch command[0] {
			case "set":
				command = command[1:]
				switch command[0] {
				case "osc":
					if len(command) != 2 {
						panic(fmt.Errorf("invalid key-value pair %v", command))
					}
					s.osc.Set(command[0], command[1])
				}
				s.osc.Set("kind", command[1])
			case "note_on":
				note, err := strconv.ParseInt(command[1], 10, 32)
				if err != nil {
					panic(err)
				}
				s.osc.SetFreq(442 * math.Pow(2, float64(note-69)/12))
				s.gain = 0.3
			case "note_off":
				s.gain = 0
			default:
				panic(fmt.Errorf("unknown command %v", command[0]))
			}
		default:
		}
	}
}
