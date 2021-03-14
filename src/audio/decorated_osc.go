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

func newDecoratedOsc() *decoratedOsc {
	return &decoratedOsc{
		oscs:       []*osc{newOsc(true), newOsc(false)},
		adsr:       &adsr{tvalue: &transitiveValue{}},
		noteFilter: newNoteFilter(),
		filter:     newFilter(),
		formant:    newFormant(),
		lfos:       []*lfo{newLfo(), newLfo(), newLfo()},
		envelopes:  []*envelope{newEnvelope(), newEnvelope(), newEnvelope()},
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
	for _, envelope := range o.envelopes {
		if !envelope.enabled {
			continue
		}
		envelope.step()
	}
	osc0VolumeRatio := 1.0
	osc1VolumeRatio := 1.0
	freqRatio := 1.0
	phaseShift := 0.0
	ampRatio := 1.0
	noteFilterFreqRatio := 1.0
	noteFilterQExponent := 1.0
	noteFilterGainRatio := 1.0
	filterFreqRatio := 1.0
	filterQExponent := 1.0
	filterGainRatio := 1.0
	for lfoIndex, lfo := range o.lfos {
		amountGain := 1.0
		lfoFreqRatio := 1.0
		for _, envelope := range o.envelopes {
			if !envelope.enabled {
				continue
			}
			v := envelope.getValue()
			if envelope.kind == envelopeKindGoing {
				v = 1 - v
			}
			if envelope.destination == destLfoAmount[lfoIndex] {
				amountGain *= 1 - v
			} else if envelope.destination == destLfoFreq[lfoIndex] {
				lfoFreqRatio *= math.Pow(2.0, v)
			}
		}
		_freqRatio, _phaseShift, _ampRatio, _noteFilterFreqRatio, _filterFreqRatio := lfo.step(o.oscs[0], amountGain, lfoFreqRatio) // TODO
		freqRatio *= _freqRatio
		phaseShift += _phaseShift
		ampRatio *= _ampRatio
		noteFilterFreqRatio *= _noteFilterFreqRatio
		filterFreqRatio *= _filterFreqRatio
	}
	for _, envelope := range o.envelopes {
		if !envelope.enabled {
			continue
		}
		v := envelope.getValue()
		if envelope.kind == envelopeKindGoing {
			v = 1 - v
		}
		if envelope.destination == destOsc0Volume {
			osc0VolumeRatio *= 1 - v
		} else if envelope.destination == destOsc1Volume {
			osc1VolumeRatio *= 1 - v
		} else if envelope.destination == destFreq {
			freqRatio *= math.Pow(2.0, v*envelope.amount)
		} else if envelope.destination == destNoteFilterFreq {
			noteFilterFreqRatio *= math.Pow(2.0, v*envelope.amount)
		} else if envelope.destination == destNoteFilterQ {
			noteFilterQExponent *= 1 - v
		} else if envelope.destination == destNoteFilterGain {
			noteFilterGainRatio *= 1 - v
		} else if envelope.destination == destFilterFreq {
			filterFreqRatio *= math.Pow(2.0, v*envelope.amount)
		} else if envelope.destination == destFilterQ {
			filterQExponent *= 1 - v
		} else if envelope.destination == destFilterGain {
			filterGainRatio *= 1 - v
		}
	}
	value := 0.0
	for i, osc := range o.oscs {
		v := osc.step(freqRatio, phaseShift) * oscGain * ampRatio * o.adsr.getValue()
		if i == 0 {
			v *= osc0VolumeRatio
		} else if i == 1 {
			v *= osc1VolumeRatio
		}
		if i == 0 && o.noteFilter.targetOsc == targetOsc0 ||
			i == 1 && o.noteFilter.targetOsc == targetOsc1 {
			v = o.noteFilter.step(v, noteFilterFreqRatio, noteFilterQExponent, noteFilterGainRatio, o.oscs[0].freq.value*freqRatio) // TODO: use original freq of note
		}
		if i == 0 && o.filter.targetOsc == targetOsc0 ||
			i == 1 && o.filter.targetOsc == targetOsc1 {
			v = o.filter.step(v, filterFreqRatio, filterQExponent, filterGainRatio)
		}
		value += v
	}
	if o.noteFilter.targetOsc == targetOscAll {
		value = o.noteFilter.step(value, filterFreqRatio, noteFilterQExponent, noteFilterGainRatio, o.oscs[0].freq.value*freqRatio) // TODO: use original freq of note
	}
	if o.filter.targetOsc == targetOscAll {
		value = o.filter.step(value, filterFreqRatio, filterQExponent, filterGainRatio)
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
