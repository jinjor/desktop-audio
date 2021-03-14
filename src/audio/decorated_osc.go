package audio

import (
	"math"
)

// ----- Decorated OSC -----

type decoratedOsc struct {
	oscs       []*osc
	adsr       *adsr
	noteFilter *noteFilter
	filter     *filter
	formant    *formant
	lfos       []*lfo
	envelopes  []*envelope
	modulation *modulation
}

func newDecoratedOsc() *decoratedOsc {
	return &decoratedOsc{
		oscs:       []*osc{newOsc(true), newOsc(false)},
		adsr:       &adsr{tvalue: &transitiveValue{}},
		noteFilter: newNoteFilter(),
		filter:     newFilter(),
		formant:    newFormant(),
		lfos:       []*lfo{newLfo(), newLfo(), newLfo()},
		envelopes:  []*envelope{newEnvelope(), newEnvelope(), newEnvelope()},
		modulation: newModulation(),
	}
}

const (
	enumNoEvent = iota
	enumNoteOn
	enumNoteOff
)

func (o *decoratedOsc) initWithNote(p []*oscParams, note int) {
	for i, osc := range o.oscs {
		osc.initWithNote(p[i], note)
	}
}
func (o *decoratedOsc) glide(p []*oscParams, note int, glideTime int) {
	for i, osc := range o.oscs {
		osc.glide(p[i], note, glideTime)
	}
}
func (o *decoratedOsc) applyParams(
	oscParams []*oscParams,
	adsrParams *adsrParams,
	noteFilterParams *noteFilterParams,
	filterParams *filterParams,
	formantParams *formantParams,
	lfoParams []*lfoParams,
	envelopeParams []*envelopeParams,
) {
	o.adsr.setParams(adsrParams)
	o.noteFilter.applyParams(noteFilterParams)
	o.filter.applyParams(filterParams)
	o.formant.applyParams(formantParams)
	for i, lfo := range o.lfos {
		lfo.applyParams(lfoParams[i])
	}
	for i, envelope := range o.envelopes {
		envelope.enabled = envelopeParams[i].enabled
		envelope.destination = envelopeParams[i].destination
		envelope.applyParams(envelopeParams[i])
	}
}

func (o *decoratedOsc) step(event int) float64 {
	switch event {
	case enumNoEvent:
	case enumNoteOn:
		o.adsr.noteOn()
		for _, envelope := range o.envelopes {
			envelope.noteOn()
		}
	case enumNoteOff:
		o.adsr.noteOff()
		for _, envelope := range o.envelopes {
			envelope.noteOff()
		}
	}
	o.adsr.step()
	m := o.modulation
	m.init()
	for _, envelope := range o.envelopes {
		envelope.step(m)
	}
	for lfoIndex, lfo := range o.lfos {
		lfo.step(o.oscs[0], m.lfoAmountGain[lfoIndex], m.lfoFreqRatio[lfoIndex], m)
	}
	value := 0.0
	for i, osc := range o.oscs {
		v := osc.step(m.freqRatio, m.phaseShift) * oscGain * m.ampRatio * o.adsr.getValue()
		v *= m.oscVolumeRatio[i]
		if i == 0 && o.noteFilter.targetOsc == targetOsc0 ||
			i == 1 && o.noteFilter.targetOsc == targetOsc1 {
			v = o.noteFilter.step(v, m.noteFilterFreqRatio, m.noteFilterQExponent, m.noteFilterGainRatio, o.oscs[0].freq.value*m.freqRatio) // TODO: use original freq of note
		}
		if i == 0 && o.filter.targetOsc == targetOsc0 ||
			i == 1 && o.filter.targetOsc == targetOsc1 {
			v = o.filter.step(v, m.filterFreqRatio, m.filterQExponent, m.filterGainRatio)
		}
		value += v
	}
	if o.noteFilter.targetOsc == targetOscAll {
		value = o.noteFilter.step(value, m.filterFreqRatio, m.noteFilterQExponent, m.noteFilterGainRatio, o.oscs[0].freq.value*m.freqRatio) // TODO: use original freq of note
	}
	if o.filter.targetOsc == targetOscAll {
		value = o.filter.step(value, m.filterFreqRatio, m.filterQExponent, m.filterGainRatio)
	}
	value = o.formant.step(value)
	if math.IsNaN(value) {
		panic("found NaN")
	}
	if math.IsInf(value, 0) {
		panic("found NaN")
	}
	return value
}
