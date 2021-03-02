package audio

import (
	"encoding/json"
	"log"
	"math"
	"strconv"
)

// ----- ADSR Params ----- //

const (
	phaseNone = iota
	phaseAttack
	phaseKeep
	phaseDecay
	phaseSustain
	phaseRelease
)

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

// ----- ADSR ----- //

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
	phase          int // "attack", "keep", "decay", "sustain", "release" nil
	phasePos       int
	valueAtNoteOn  float64
	valueAtNoteOff float64
}

func (a *adsr) init(p *adsrParams) {
	a.setParams(p)
	a.value = 0
	a.phase = phaseNone
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
	if p.destination == destVibrato ||
		p.destination == destTremolo ||
		p.destination == destFilterQ0V ||
		p.destination == destFilterGain0V ||
		p.destination == destLfo0Amount ||
		p.destination == destLfo1Amount ||
		p.destination == destLfo2Amount {
		// zero-to-value
		a.base = 1
		a.peek = 0
	} else if p.destination == destFreq ||
		p.destination == destFilterFreq ||
		p.destination == destLfo0Freq ||
		p.destination == destLfo1Freq ||
		p.destination == destLfo2Freq {
		// amount-to-value
		a.base = 0
		a.peek = p.amount
	} else if p.destination == destFilterQ ||
		p.destination == destFilterGain {
		// value-to-zero
		a.base = 0
		a.peek = 1
	}
	a.attack = 0
	a.keep = p.delay
	a.decay = p.attack
	a.sustain = a.base
	a.release = 0
}

func (a *adsr) noteOn() {
	a.phase = phaseAttack
	a.phasePos = 0
	a.valueAtNoteOn = a.value
}

func (a *adsr) noteOff() {
	a.phase = phaseRelease
	a.phasePos = 0
	a.valueAtNoteOff = a.value
}

func (a *adsr) step() {
	phaseTime := float64(a.phasePos) * secPerSample * 1000 // ms
	switch a.phase {
	case phaseAttack:
		if phaseTime >= float64(a.attack) {
			a.phase = phaseKeep
			a.phasePos = 0
			a.value = a.peek
		} else {
			t := phaseTime / float64(a.attack)
			a.value = t*a.peek + (1-t)*a.valueAtNoteOn // TODO: don't use the same attack time
			a.phasePos++
		}
	case phaseKeep:
		if phaseTime >= float64(a.keep) {
			a.phase = phaseDecay
			a.phasePos = 0
		} else {
			a.phasePos++
		}
	case phaseDecay:
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
			a.phase = phaseSustain
			a.phasePos = 0
			a.value = float64(a.sustain)
		} else {
			a.phasePos++
		}
	case phaseSustain:
		a.value = float64(a.sustain)
	case phaseRelease:
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
			a.phase = phaseNone
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
