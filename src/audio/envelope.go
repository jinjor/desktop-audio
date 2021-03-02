package audio

import (
	"encoding/json"
	"log"
	"strconv"
)

// ----- Envelope ----- //

type envelopeParams struct {
	enabled     bool
	destination int
	delay       float64
	attack      float64
	amount      float64
}

type envelopeJSON struct {
	Enabled     bool    `json:"enabled"`
	Destination string  `json:"destination"`
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
	l.delay = j.Delay
	l.attack = j.Attack
	l.amount = j.Amount
}
func (l *envelopeParams) toJSON() json.RawMessage {
	return toRawMessage(&envelopeJSON{
		Enabled:     l.enabled,
		Destination: destinationToString(l.destination),
		Delay:       l.delay,
		Attack:      l.attack,
		Amount:      l.amount,
	})
}
func newEnvelopeParams() *envelopeParams {
	return &envelopeParams{
		enabled:     false,
		destination: destNone,
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

/*
  [value-to-zero] filter_q, etc.
  v +_
	  | \_
    |   \_
    |     \_
  0 +-------+---------
    |attack |

  [amount-to-value] freq, filter_freq, etc.
  a +_
	  | \_
    |   \_
    |     \_
  v +       `-----
	  |
		+--------+--------
    |attack  |

  [zero-to-value] vibrato, tremolo, etc.
  v +           ,-----
	  |         _/
    |       _/
    |      /
  0 +-----+------+----
    |delay|attack|
*/
type envelope struct {
	enabled bool
	*adsr
	destination int
}

func newEnvelope() *envelope {
	return &envelope{
		enabled:     false,
		destination: destNone,
		adsr:        &adsr{},
	}
}
