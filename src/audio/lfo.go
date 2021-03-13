package audio

import (
	"encoding/json"
	"log"
	"math"
	"strconv"
)

// ----- LFO Params ----- //

type lfoParams struct {
	enabled     bool
	destination int
	wave        int
	freqType    string
	freq        float64
	amount      float64
}

type lfoJSON struct {
	Enabled     bool    `json:"enabled"`
	Destination string  `json:"destination"`
	Wave        string  `json:"wave"`
	FreqType    string  `json:"freqType"`
	Freq        float64 `json:"freq"`
	Amount      float64 `json:"amount"`
}

func (l *lfoParams) applyJSON(data json.RawMessage) {
	var j lfoJSON
	err := json.Unmarshal(data, &j)
	if err != nil {
		log.Println("failed to apply JSON to adsrParams")
		return
	}
	l.enabled = j.Enabled
	l.destination = destinationFromString(j.Destination)
	l.wave = waveKindFromString(j.Wave)
	l.freqType = j.FreqType
	l.freq = j.Freq
	l.amount = j.Amount
}
func (l *lfoParams) toJSON() json.RawMessage {
	return toRawMessage(&lfoJSON{
		Enabled:     l.enabled,
		Destination: destinationToString(l.destination),
		Wave:        waveKindToString(l.wave),
		FreqType:    l.freqType,
		Freq:        l.freq,
		Amount:      l.amount,
	})
}

func newLfoParams() *lfoParams {
	return &lfoParams{
		enabled:     false,
		destination: destNone,
		wave:        waveSine,
		freqType:    "none",
		freq:        0,
		amount:      0,
	}
}

// func (l *lfoParams) initByDestination(destination string) {
// 	l.destination = destination
// 	switch destination {
// 	case "vibrato":
// 		l.freqType = "absolute"
// 		l.freq = 0
// 		l.amount = 0
// 	case "tremolo":
// 		l.freqType = "absolute"
// 		l.freq = 0
// 		l.amount = 0
// 	}
// }

func (l *lfoParams) set(key string, value string) error {
	switch key {
	case "enabled":
		// l.initByDestination(value)
		l.enabled = value == "true"
	case "destination":
		// l.initByDestination(value)
		l.destination = destinationFromString(value)
	case "wave":
		l.wave = waveKindFromString(value)
	case "freq":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		l.freq = value
	case "amount":
		value, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		l.amount = value
	}
	return nil
}

// ----- LFO ----- //

type lfo struct {
	enabled     bool
	destination int
	freqType    string
	amount      float64
	osc         *osc
}

func newLfo() *lfo {
	return &lfo{
		enabled:     false,
		destination: destNone,
		freqType:    "none",
		amount:      0,
		osc:         newOsc(true),
	}
}

func (l *lfo) applyParams(p *lfoParams) {
	l.enabled = p.enabled
	l.destination = p.destination
	l.osc.kind = p.wave // TODO
	l.freqType = p.freqType
	l.osc.freq.value = p.freq // TODO
	l.amount = p.amount
}

func (l *lfo) step(career *osc, amountGain float64, lfoFreqRatio float64) (float64, float64, float64, float64, float64) {
	freqRatio := 1.0
	phaseShift := 0.0
	ampRatio := 1.0
	noteFilterFreqRatio := 1.0
	filterFreqRatio := 1.0
	if !l.enabled {
		return freqRatio, phaseShift, ampRatio, noteFilterFreqRatio, filterFreqRatio
	}
	switch l.destination {
	case destVibrato:
		amount := l.amount * amountGain
		freqRatio = math.Pow(2.0, l.osc.step(lfoFreqRatio, 0.0)*amount/100.0/12.0)
	case destTremolo:
		amount := l.amount * amountGain
		ampRatio = 1.0 + (l.osc.step(lfoFreqRatio, 0.0)-1.0)/2.0*amount
	case destFM:
		amount := l.amount * amountGain
		freqRatio = math.Pow(2.0, l.osc.step(career.freq.value*lfoFreqRatio, 0.0)*amount/100/12)
	case destPM:
		amount := l.amount * amountGain
		phaseShift = l.osc.step(career.freq.value*lfoFreqRatio, 0.0) * amount
	case destAM:
		amount := l.amount * amountGain
		ampRatio = 1.0 + l.osc.step(career.freq.value*lfoFreqRatio, 0.0)*amount
	case destNoteFilterFreq:
		amount := l.amount * amountGain
		noteFilterFreqRatio = math.Pow(16.0, l.osc.step(lfoFreqRatio, 0.0)*amount)
	case destFilterFreq:
		amount := l.amount * amountGain
		filterFreqRatio = math.Pow(16.0, l.osc.step(lfoFreqRatio, 0.0)*amount)
	}
	return freqRatio, phaseShift, ampRatio, noteFilterFreqRatio, filterFreqRatio
}
