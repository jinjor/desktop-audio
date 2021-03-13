package audio

import (
	"encoding/json"
	"log"
)

type params struct {
	polyMode         bool
	glideTime        int     // ms
	velSense         float64 // 0-1
	oscParams        []*oscParams
	adsrParams       *adsrParams
	noteFilterParams *noteFilterParams
	filterParams     *filterParams
	formantParams    *formantParams
	lfoParams        []*lfoParams
	envelopeParams   []*envelopeParams
	echoParams       *echoParams
}

func newParams() *params {
	return &params{
		oscParams:        []*oscParams{{enabled: true, kind: waveSine, level: 1.0}, {enabled: false, kind: waveSine, level: 1.0}},
		adsrParams:       &adsrParams{attack: 10, decay: 100, sustain: 0.7, release: 200},
		lfoParams:        []*lfoParams{newLfoParams(), newLfoParams(), newLfoParams()},
		noteFilterParams: &noteFilterParams{kind: filterNone, q: 1, gain: 0},
		filterParams:     &filterParams{kind: filterNone, freq: 1000, q: 1, gain: 0, N: 50},
		formantParams:    &formantParams{kind: formantA, tone: 1, q: 1},
		envelopeParams:   []*envelopeParams{newEnvelopeParams(), newEnvelopeParams(), newEnvelopeParams()},
		echoParams:       &echoParams{},
		polyMode:         false,
		glideTime:        100,
		velSense:         0,
	}
}

type paramsJSON struct {
	Poly       string            `json:"poly"`
	GlideTime  int               `json:"glideTime"`
	VelSense   float64           `json:"velSense"`
	Oscs       []json.RawMessage `json:"oscs"`
	Adsr       json.RawMessage   `json:"adsr"`
	NoteFilter json.RawMessage   `json:"noteFilter"`
	Filter     json.RawMessage   `json:"filter"`
	Formant    json.RawMessage   `json:"formant"`
	Lfos       []json.RawMessage `json:"lfos"`
	Envelopes  []json.RawMessage `json:"envelopes"`
	Echo       json.RawMessage   `json:"echo"`
}

func (p *params) applyJSON(data json.RawMessage) {
	var j paramsJSON
	err := json.Unmarshal(data, &j)
	if err != nil {
		log.Println(err)
		log.Println(data)
		log.Println("failed to apply JSON to state")
		return
	}
	p.polyMode = j.Poly == "poly"
	p.glideTime = j.GlideTime
	p.velSense = j.VelSense
	if len(j.Oscs) == len(p.oscParams) {
		for i, j := range j.Oscs {
			p.oscParams[i].applyJSON(j)
		}
	} else {
		log.Println("failed to apply JSON to osc params")
	}
	p.adsrParams.applyJSON(j.Adsr)
	p.noteFilterParams.applyJSON(j.NoteFilter)
	p.filterParams.applyJSON(j.Filter)
	p.formantParams.applyJSON(j.Formant)
	if len(j.Lfos) == len(p.lfoParams) {
		for i, j := range j.Lfos {
			p.lfoParams[i].applyJSON(j)
		}
	} else {
		log.Println("failed to apply JSON to lfo params")
	}
	if len(j.Envelopes) == len(p.envelopeParams) {
		for i, j := range j.Envelopes {
			p.envelopeParams[i].applyJSON(j)
		}
	} else {
		log.Println("failed to apply JSON to envelope params")
	}
	p.echoParams.applyJSON(j.Echo)
}
func (p *params) toJSON() json.RawMessage {
	oscJsons := make([]json.RawMessage, len(p.oscParams))
	for i, oscParam := range p.oscParams {
		oscJsons[i] = oscParam.toJSON()
	}
	lfoJsons := make([]json.RawMessage, len(p.lfoParams))
	for i, lfoParam := range p.lfoParams {
		lfoJsons[i] = lfoParam.toJSON()
	}
	envelopeJsons := make([]json.RawMessage, len(p.envelopeParams))
	for i, envelopeParam := range p.envelopeParams {
		envelopeJsons[i] = envelopeParam.toJSON()
	}
	poly := "mono"
	if p.polyMode {
		poly = "poly"
	}
	return toRawMessage(&paramsJSON{
		Poly:       poly,
		GlideTime:  p.glideTime,
		VelSense:   p.velSense,
		Oscs:       oscJsons,
		Adsr:       p.adsrParams.toJSON(),
		NoteFilter: p.noteFilterParams.toJSON(),
		Filter:     p.filterParams.toJSON(),
		Formant:    p.formantParams.toJSON(),
		Lfos:       lfoJsons,
		Envelopes:  envelopeJsons,
		Echo:       p.echoParams.toJSON(),
	})
}
