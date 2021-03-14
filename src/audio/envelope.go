package audio

import (
	"encoding/json"
	"log"
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
func (a *envelope) applyParams(p *envelopeParams) {
	a.destination = p.destination
	a.kind = p.kind
	a.amount = p.amount
	a.base = 0
	a.peak = 1
	a.attack = 0
	a.hold = p.delay
	a.decay = p.attack
	a.sustain = a.base
	a.release = 0
	if a.tvalue == nil {
		a.tvalue = &transitiveValue{}
	}
	a.tvalue.value = a.base
}
func (a *envelope) noteOff() {
	// noop
}
