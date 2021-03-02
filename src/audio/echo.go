package audio

import (
	"encoding/json"
	"log"
	"strconv"
)

// ----- Delay ----- //

type delay struct {
	cursor int
	past   []float64
}

func (d *delay) applyParams(millis float64) {
	if millis < 10 {
		millis = 10
	}
	length := int(sampleRate * millis / 1000)
	if cap(d.past) >= length {
		d.past = d.past[0:length]
	} else {
		d.past = make([]float64, length)
	}
	if d.cursor >= len(d.past) {
		d.cursor = 0
	}
}

func (d *delay) step(in float64) {
	d.past[d.cursor] = in
	d.cursor++
	if d.cursor >= len(d.past) {
		d.cursor = 0
	}
}
func (d *delay) getDelayed() float64 {
	return d.past[d.cursor]
}

// ----- Echo ----- //

type echoParams struct {
	enabled      bool
	delay        float64
	feedbackGain float64
	mix          float64
}

type echoJSON struct {
	Enabled      bool    `json:"enabled"`
	Delay        float64 `json:"delay"`
	FeedbackGain float64 `json:"feedbackGain"`
	Mix          float64 `json:"mix"`
}

func (l *echoParams) applyJSON(data json.RawMessage) {
	var j echoJSON
	err := json.Unmarshal(data, &j)
	if err != nil {
		log.Println("failed to apply JSON to echoParams")
		return
	}
	l.enabled = j.Enabled
	l.delay = j.Delay
	l.feedbackGain = j.FeedbackGain
	l.mix = j.Mix
}
func (l *echoParams) toJSON() json.RawMessage {
	return toRawMessage(&echoJSON{
		Enabled:      l.enabled,
		Delay:        l.delay,
		FeedbackGain: l.feedbackGain,
		Mix:          l.mix,
	})
}
func (l *echoParams) set(key string, value string) error {
	switch key {
	case "enabled":
		l.enabled = value == "true"
	case "delay":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		l.delay = value
	case "feedbackGain":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		l.feedbackGain = value
	case "mix":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		l.mix = value
	}
	return nil
}

type echo struct {
	enabled      bool
	delay        *delay
	feedbackGain float64 // [0,1)
	mix          float64 // [0,1]
}

func (e *echo) applyParams(p *echoParams) {
	e.enabled = p.enabled
	e.delay.applyParams(p.delay)
	e.feedbackGain = p.feedbackGain
	e.mix = p.mix
}

func (e *echo) step(in float64) float64 {
	if !e.enabled {
		return in
	}
	delayed := e.delay.getDelayed()
	e.delay.step(in + delayed*e.feedbackGain)
	return in + delayed*e.mix
}
