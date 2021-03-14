package audio

import (
	"encoding/json"
	"log"
	"strconv"
)

// ----- ADSR Params ----- //

const (
	phaseNone = iota
	phaseAttack
	phaseHold
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
    |a    |h  |d |      |r |
*/
type adsr struct {
	attack  float64 // ms
	hold    float64 // ms
	decay   float64 // ms
	sustain float64 // 0-1
	release float64 // ms
	base    float64 // 0-1
	peak    float64 // 0-1
	phase   int     // "none", "attack", "hold", "decay", "sustain", "release"
	tvalue  *transitiveValue
}

func (a *adsr) getValue() float64 {
	return a.tvalue.value
}
func (a *adsr) setParams(p *adsrParams) {
	a.base = 0
	a.peak = 1
	a.attack = p.attack
	a.hold = 0
	a.decay = p.decay
	a.sustain = p.sustain
	a.release = p.release
	if a.tvalue == nil {
		a.tvalue = &transitiveValue{}
	}
}
func (a *adsr) noteOn() {
	a.phase = phaseAttack
	a.tvalue.linear(a.attack, a.peak)
}
func (a *adsr) noteOff() {
	a.phase = phaseRelease
	a.tvalue.exponential(a.release, a.base, 0.001)
}
func (a *adsr) step() {
	switch a.phase {
	case phaseAttack:
		if a.tvalue.step() {
			a.phase = phaseHold
			a.tvalue.linear(a.hold, a.peak)
		}
	case phaseHold:
		if a.tvalue.step() {
			a.phase = phaseDecay
			a.tvalue.exponential(a.decay, a.sustain, 0.001)
		}
	case phaseDecay:
		if a.tvalue.step() {
			a.phase = phaseSustain
		}
	case phaseSustain:
	case phaseRelease:
		if a.tvalue.step() {
			a.phase = phaseNone
		}
	default:
	}
}
