package audio

import (
	"encoding/json"
	"log"
	"strconv"
)

// ----- Formant Kind ----- //

//go:generate go run ../gen/main.go -- formant_kind.gen.go
/*
generate-enum formantKind

formantA a
formantE e
formantI i
formantO o
formantU u

EOF
*/

// ----- Formant Params ----- //

type formantParams struct {
	enabled bool
	kind    int
	tone    float64
	q       float64
}

type formantJSON struct {
	Enabled bool    `json:"enabled"`
	Kind    string  `json:"kind"`
	Tone    float64 `json:"tone"`
	Q       float64 `json:"q"`
}

func (f *formantParams) applyJSON(data json.RawMessage) {
	var j formantJSON
	err := json.Unmarshal(data, &j)
	if err != nil {
		log.Println("failed to apply JSON to formant")
		return
	}
	f.enabled = j.Enabled
	f.kind = formantKindFromString(j.Kind)
	f.tone = j.Tone
	f.q = j.Q
}
func (f *formantParams) toJSON() json.RawMessage {
	return toRawMessage(&formantJSON{
		Enabled: f.enabled,
		Kind:    formantKindToString(f.kind),
		Tone:    f.tone,
		Q:       f.q,
	})
}
func (f *formantParams) set(key string, value string) error {
	switch key {
	case "enabled":
		f.enabled = value == "true"
	case "kind":
		f.kind = formantKindFromString(value)
	case "freq":
		tone, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		f.tone = tone
	case "q":
		q, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		f.q = q
	}
	return nil
}

// ----- Formant ----- //

var formantFreqs = initFormantFreqs()

func initFormantFreqs() [][]float64 {
	freqs := make([][]float64, 5)
	for range freqs {
		freqs[formantA] = []float64{800, 1200, 2500, 3500}
		freqs[formantE] = []float64{500, 1900, 2500, 3500}
		freqs[formantI] = []float64{300, 2300, 2900, 3500}
		freqs[formantO] = []float64{500, 800, 2500, 3500}
		freqs[formantU] = []float64{300, 1200, 2500, 3500}
	}
	return freqs
}

type formant struct {
	enabled bool
	kind    int
	tone    float64
	filters []*filter
}

func newFormant() *formant {
	kind := formantA
	formant := &formant{
		enabled: false,
		kind:    kind,
		tone:    1,
		filters: make([]*filter, 4),
	}
	formant.applyFreqs(kind)
	formant.applyQ(1)
	return formant
}
func (f *formant) applyFreqs(kind int) {
	freqs := formantFreqs[kind]
	for i, filter := range f.filters {
		params := &filterParams{
			enabled: true,
			kind:    filterBandPass1,
			freq:    freqs[i],
			q:       1,
			gain:    0,
		}
		filter.applyParams(params)
	}
}
func (f *formant) applyQ(q float64) {
	for _, filter := range f.filters {
		filter.q = q // TODO: filter.applyQ
	}
}
func (f *formant) applyParams(p *formantParams) {
	f.enabled = p.enabled
	if f.kind != p.kind {
		f.kind = p.kind
		f.applyFreqs(p.kind)
	}
	f.applyQ(p.q)
	f.tone = p.tone
}
func (f *formant) step(in float64, freqRatio float64) float64 {
	if !f.enabled {
		return in
	}
	out := 0.0
	for _, filter := range f.filters {
		out += filter.step(in, f.tone, nil)
	}
	return out
}
