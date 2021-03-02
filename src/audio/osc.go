package audio

import (
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
)

// ----- Wave Kind ----- //

//go:generate go run ../gen/main.go -- wave_kind.gen.go
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

// ----- OSC Params ----- //

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

// ----- OSC ----- //

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

var blsquareWT *WavetableSet = loadWavetableSet("square")
var blsawWT *WavetableSet = loadWavetableSet("saw")

func loadWavetableSet(path string) *WavetableSet {
	wts := NewWavetableSet(128, 4096)
	if os.Getenv("AUDIO_TESTING") == "1" {
		path = "../../work/" + path
	} else {
		path = "work/" + path
	}
	err := wts.Load(path)
	if err != nil && os.Getenv("WAVETABLE_GENERATION") != "1" {
		panic(err)
	}
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
