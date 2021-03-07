package audio

import (
	"math"
)

type decoratedOsc struct {
	oscs       []*osc
	adsr       *adsr
	noteFilter *noteFilter
	filter     *filter
	formant    *formant
	lfos       []*lfo
	envelopes  []*envelope
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
		envelope.adsr.applyEnvelopeParams(envelopeParams[i])
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
	for _, envelope := range o.envelopes {
		if !envelope.enabled {
			continue
		}
		envelope.step()
	}
	freqRatio := 1.0
	phaseShift := 0.0
	ampRatio := 1.0
	filterFreqRatio := 1.0
	for lfoIndex, lfo := range o.lfos {
		amountGain := 1.0
		lfoFreqRatio := 1.0
		for _, envelope := range o.envelopes {
			if !envelope.enabled {
				continue
			}
			if envelope.destination == destLfoAmount[lfoIndex] {
				amountGain *= envelope.getValue()
			}
			if envelope.destination == destLfoFreq[lfoIndex] {
				lfoFreqRatio *= math.Pow(16.0, envelope.getValue())
			}
		}
		_freqRatio, _phaseShift, _ampRatio, _filterFreqRatio := lfo.step(o.oscs[0], amountGain, lfoFreqRatio) // TODO
		freqRatio *= _freqRatio
		phaseShift += _phaseShift
		ampRatio *= _ampRatio
		filterFreqRatio *= _filterFreqRatio
	}
	for _, envelope := range o.envelopes {
		if !envelope.enabled {
			continue
		}
		if envelope.destination == destFreq {
			freqRatio *= math.Pow(16.0, envelope.getValue())
		}
	}
	value := 0.0
	for _, osc := range o.oscs {
		value += osc.step(freqRatio, phaseShift) * oscGain * ampRatio * o.adsr.getValue()
	}
	value = o.noteFilter.step(value, filterFreqRatio, o.envelopes, o.oscs[o.noteFilter.baseOsc].freq.value*freqRatio)
	value = o.filter.step(value, filterFreqRatio, o.envelopes)
	value = o.formant.step(value)
	if math.IsNaN(value) {
		panic("found NaN")
	}
	if math.IsInf(value, 0) {
		panic("found NaN")
	}
	return value
}
