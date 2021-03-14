package audio

import (
	"encoding/json"
	"log"
	"math"
	"strconv"
)

// ----- Envelope Kind ----- //

//go:generate go run ../gen/main.go -- envelopeKind.gen.go
/*
generate-enum envelopeKind

envelopeKindComing coming
envelopeKindGoing going

EOF

  [coming]
        ? +______
	        |      \_
          |        \_
          |          \_
  setting + - - - - - -`----
	        |
		      +-----+------+----
          |delay|attack|

  [going]
        ? +           ,-----
	        |         _/
          |       _/
          |     _/
  setting +----- - - - - - -
		      |
          +-----+------+----
          |delay|attack|
*/

// ----- Envelope ----- //

type envelopeParams struct {
	enabled     bool
	destination int
	kind        int
	delay       float64
	attack      float64
	amount      float64
}

type envelopeJSON struct {
	Enabled     bool    `json:"enabled"`
	Destination string  `json:"destination"`
	Kind        string  `json:"kind"`
	Delay       float64 `json:"delay"`
	Attack      float64 `json:"attack"`
	Amount      float64 `json:"amount"`
}

func (l *envelopeParams) applyJSON(data json.RawMessage) {
	var j envelopeJSON
	err := json.Unmarshal(data, &j)
	if err != nil {
		log.Println("failed to apply JSON to envelopeParams")
		return
	}
	l.enabled = j.Enabled
	l.destination = destinationFromString(j.Destination)
	l.kind = envelopeKindFromString(j.Kind)
	l.delay = j.Delay
	l.attack = j.Attack
	l.amount = j.Amount
}
func (l *envelopeParams) toJSON() json.RawMessage {
	return toRawMessage(&envelopeJSON{
		Enabled:     l.enabled,
		Destination: destinationToString(l.destination),
		Kind:        envelopeKindToString(l.kind),
		Delay:       l.delay,
		Attack:      l.attack,
		Amount:      l.amount,
	})
}
func newEnvelopeParams() *envelopeParams {
	return &envelopeParams{
		enabled:     false,
		destination: destNone,
		kind:        envelopeKindComing,
		delay:       0,
		attack:      0,
		amount:      0,
	}
}
func (l *envelopeParams) set(key string, value string) error {
	switch key {
	case "enabled":
		l.enabled = value == "true"
	case "destination":
		l.destination = destinationFromString(value)
	case "kind":
		l.kind = envelopeKindFromString(value)
	case "delay":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		l.delay = value
	case "attack":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		l.attack = value
	case "amount":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		l.amount = value
	}
	return nil
}

type envelope struct {
	*adsr
	enabled     bool
	destination int
	kind        int
	amount      float64
}

func newEnvelope() *envelope {
	return &envelope{
		adsr:        &adsr{tvalue: &transitiveValue{}},
		enabled:     false,
		destination: destNone,
		kind:        envelopeKindComing,
		amount:      0,
	}
}
func (e *envelope) applyParams(p *envelopeParams) {
	e.destination = p.destination
	e.kind = p.kind
	e.amount = p.amount
	e.base = 0
	e.peak = 1
	e.attack = 0
	e.hold = p.delay
	e.decay = p.attack
	e.sustain = e.base
	e.release = 0
	if e.tvalue == nil {
		e.tvalue = &transitiveValue{}
	}
	e.tvalue.value = e.base
}
func (e *envelope) noteOff() {
	// noop
}
func (e *envelope) step(m *modulation) {
	if !e.enabled {
		return
	}
	e.adsr.step()
	v := e.adsr.getValue()
	if e.kind == envelopeKindGoing {
		v = 1 - v
	}
	if e.destination == destOsc0Volume {
		m.oscVolumeRatio[0] *= 1 - v
	} else if e.destination == destOsc1Volume {
		m.oscVolumeRatio[1] *= 1 - v
	} else if e.destination == destFreq {
		m.freqRatio *= math.Pow(2.0, v*e.amount)
	} else if e.destination == destNoteFilterFreq {
		m.noteFilterFreqRatio *= math.Pow(2.0, v*e.amount)
	} else if e.destination == destNoteFilterQ {
		m.noteFilterQExponent *= 1 - v
	} else if e.destination == destNoteFilterGain {
		m.noteFilterGainRatio *= 1 - v
	} else if e.destination == destFilterFreq {
		m.filterFreqRatio *= math.Pow(2.0, v*e.amount)
	} else if e.destination == destFilterQ {
		m.filterQExponent *= 1 - v
	} else if e.destination == destFilterGain {
		m.filterGainRatio *= 1 - v
	} else if e.destination == destLfo0Freq {
		m.lfoFreqRatio[0] *= math.Pow(2.0, v)
	} else if e.destination == destLfo1Freq {
		m.lfoFreqRatio[1] *= math.Pow(2.0, v)
	} else if e.destination == destLfo2Freq {
		m.lfoFreqRatio[2] *= math.Pow(2.0, v)
	} else if e.destination == destLfo0Amount {
		m.lfoAmountGain[0] *= 1 - v
	} else if e.destination == destLfo1Amount {
		m.lfoAmountGain[1] *= 1 - v
	} else if e.destination == destLfo2Amount {
		m.lfoAmountGain[2] *= 1 - v
	}
}
