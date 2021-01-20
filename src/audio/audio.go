package audio

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"math/cmplx"
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
	case "lowpass":
		return makeBiquadLowpassH(fc, f.q)
	case "highpass":
		return makeBiquadHighpassH(fc, f.q)
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

func frequencyResponse(a []float64, b []float64, samples int, plots int) []float64 {
	h := impulseResponse(a, b, samples)
	Hw := make([]float64, plots)
	for i := 0; i < plots; i++ {
		fc := 0.5 / float64(plots) * float64(i)
		w := 2.0 * math.Pi * fc
		sum := complex(0, 0)
		for j := 0; j < len(h); j++ {
			sum += complex(h[j], 0) * cmplx.Exp(complex(0, -float64(j)*w))
		}
		Hw[i] = cmplx.Abs(sum)
	}
	return Hw
}

func makeNoFilterH() ([]float64, []float64) {
	return []float64{1}, []float64{}
}

func makeFIRLowpassH(N int, fc float64, windowFunc func(float64) float64) ([]float64, []float64) {
	w0 := 2 * math.Pi * fc
	if N%2 != 0 {
		log.Panicf("N should be even")
	}
	h := make([]float64, N+1)
	for i := 0; i <= N; i++ {
		n := float64(i - N/2)
		h[i] = 2 * fc * sinc(w0*n)
	}
	ApplyWindow(h, windowFunc)
	return h, []float64{}
}

func makeFIRHighpassH(N int, fc float64, windowFunc func(float64) float64) ([]float64, []float64) {
	w0 := 2 * math.Pi * fc
	if N%2 != 0 {
		log.Panicf("N should be even")
	}
	h := make([]float64, N+1)
	for i := 0; i <= N; i++ {
		n := float64(i - N/2)
		h[i] = sinc(math.Pi*n) - 2*fc*sinc(w0*n)
	}
	ApplyWindow(h, windowFunc)
	return h, []float64{}
}

func makeBiquadLowpassH(fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := (1 - math.Cos(w0)) / 2
	b1 := (1 - math.Cos(w0))
	b2 := (1 - math.Cos(w0)) / 2
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	return []float64{b0 / a0, b1 / a0, b2 / a0}, []float64{a1 / a0, a2 / a0}
}

func makeBiquadHighpassH(fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := (1 + math.Cos(w0)) / 2
	b1 := -(1 + math.Cos(w0))
	b2 := (1 + math.Cos(w0)) / 2
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	return []float64{b0 / a0, b1 / a0, b2 / a0}, []float64{a1 / a0, a2 / a0}
}

func makeBiquadBandpass1H(fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := math.Sin(w0) / 2
	b1 := 0.0
	b2 := -math.Sin(w0) / 2
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	return []float64{b0 / a0, b1 / a0, b2 / a0}, []float64{a1 / a0, a2 / a0}
}

func makeBiquadBandpass2H(fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := alpha
	b1 := 0.0
	b2 := -alpha
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	return []float64{b0 / a0, b1 / a0, b2 / a0}, []float64{a1 / a0, a2 / a0}
}

func makeBiquadNotchH(fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := 1.0
	b1 := -2 * math.Cos(w0)
	b2 := 1.0
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	return []float64{b0 / a0, b1 / a0, b2 / a0}, []float64{a1 / a0, a2 / a0}
}

func makeBiquadAllpassH(fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := 1 - alpha
	b1 := -2 * math.Cos(w0)
	b2 := 1 + alpha
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	return []float64{b0 / a0, b1 / a0, b2 / a0}, []float64{a1 / a0, a2 / a0}
}

func makeBiquadPeakingEQH(fc float64, q float64, dBgain float64) ([]float64, []float64) {
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
	return []float64{b0 / a0, b1 / a0, b2 / a0}, []float64{a1 / a0, a2 / a0}
}

func makeBiquadLowShelfH(fc float64, q float64, dBgain float64) ([]float64, []float64) {
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
	return []float64{b0 / a0, b1 / a0, b2 / a0}, []float64{a1 / a0, a2 / a0}
}

func makeBiquadHighShelfH(fc float64, q float64, dBgain float64) ([]float64, []float64) {
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
	return []float64{b0 / a0, b1 / a0, b2 / a0}, []float64{a1 / a0, a2 / a0}
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
	stateCh    chan *state
	ChangeCh   chan string
	fftResult  []float64 // length: fftSize
}

var _ io.Reader = (*Audio)(nil)

type state struct {
	osc    *osc
	filter *filter
	gain   float64
	pos    int64
	oscOut []float64 // length: samplesPerCycle
	out    []float64 // length: fftSize
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
	changeCh := make(chan string, 256)
	stateCh <- &state{
		osc:    &osc{kind: "sine", freq: 442},
		filter: &filter{kind: "none", freq: 1000, q: 1, gain: 0, N: 10},
		gain:   0,
		pos:    0,
		oscOut: make([]float64, samplesPerCycle),
		out:    make([]float64, fftSize),
	}
	audio := &Audio{
		ctx:        context.Background(),
		otoContext: otoContext,
		CommandCh:  commandCh,
		stateCh:    stateCh,
		ChangeCh:   changeCh,
		fftResult:  make([]float64, fftSize),
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
			a.ChangeCh <- "filter-shape"
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
	close(a.ChangeCh)
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

// GetFilterShape ...
func (a *Audio) GetFilterShape(ctx context.Context) []float64 {
	select {
	case <-ctx.Done():
		log.Println("GetFilterShape() interrupted.")
		return nil
	case state := <-a.stateCh:
		_a, b := getH(state.filter)
		a.stateCh <- state
		return frequencyResponse(_a, b, 2048, 64)
	}
}

// GetFFT ...
func (a *Audio) GetFFT(ctx context.Context) []float64 {
	select {
	case <-ctx.Done():
		log.Println("GetFFT() interrupted.")
		return nil
	case state := <-a.stateCh:
		// out:       | 4 | 1 | 2 | 3 |
		// offset:        ^
		// fftResult: | 1 | 2 | 3 | 4 |
		// return:    |<----->|
		offset := state.pos % fftSize
		copy(a.fftResult, state.out[offset:])
		copy(a.fftResult[fftSize-offset:], state.out[:offset])
		a.stateCh <- state
		ApplyWindow(a.fftResult, Han)
		fft.CalcAbs(a.fftResult)
		for i, value := range a.fftResult {
			a.fftResult[i] = value * 2 / fftSize
		}
		return a.fftResult[:fftSize/2]
	}
}
