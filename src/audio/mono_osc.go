package audio

import (
	"math"
	"math/rand"
)

// ----- MONO OSC ----- //

type monoOsc struct {
	o           *decoratedOsc
	activeNotes []int
}

func newMonoOsc() *monoOsc {
	return &monoOsc{
		o: &decoratedOsc{
			oscs:      []*osc{{phase: rand.Float64() * 2.0 * math.Pi}, {phase: rand.Float64() * 2.0 * math.Pi}},
			adsr:      &adsr{},
			filter:    &filter{},
			formant:   newFormant(),
			lfos:      []*lfo{newLfo(), newLfo(), newLfo()},
			envelopes: []*envelope{newEnvelope(), newEnvelope(), newEnvelope()},
		},
		activeNotes: make([]int, 0, 128),
	}
}

func (m *monoOsc) calc(
	events [][]*midiEvent,
	oscParams []*oscParams,
	adsrParams *adsrParams,
	filterParams *filterParams,
	formantParams *formantParams,
	lfoParams []*lfoParams,
	envelopeParams []*envelopeParams,
	glideTime int,
	echo *echo,
	out []float64,
) {
	m.o.applyParams(oscParams, adsrParams, filterParams, formantParams, lfoParams, envelopeParams)
	for i := int64(0); i < int64(len(out)); i++ {
		event := enumNoEvent
		for _, e := range events[i] {
			switch data := e.event.(type) {
			case *noteOn:
				if len(m.activeNotes) < cap(m.activeNotes) {
					m.activeNotes = m.activeNotes[:len(m.activeNotes)+1]
					for i := len(m.activeNotes) - 1; i >= 1; i-- {
						m.activeNotes[i] = m.activeNotes[i-1]
					}
					m.activeNotes[0] = data.note
					if len(m.activeNotes) == 1 {
						m.o.initWithNote(oscParams, data.note)
						event = enumNoteOn
					} else {
						m.o.glide(oscParams, m.activeNotes[0], glideTime)
					}
				}
			case *noteOff:
				removed := 0
				for i := 0; i < len(m.activeNotes); i++ {
					if m.activeNotes[i] == data.note {
						removed++
					} else {
						m.activeNotes[i-removed] = m.activeNotes[i]
					}
				}
				m.activeNotes = m.activeNotes[:len(m.activeNotes)-removed]
				if len(m.activeNotes) > 0 {
					m.o.glide(oscParams, m.activeNotes[0], glideTime)
				} else {
					event = enumNoteOff
				}
			}
		}
		out[i] = m.o.step(event)
		out[i] = echo.step(out[i])
	}
}
