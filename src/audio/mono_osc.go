package audio

// ----- MONO OSC ----- //

type monoOsc struct {
	o           *decoratedOsc
	activeNotes []*noteOn
	gain        *transitiveValue
}

func newMonoOsc() *monoOsc {
	return &monoOsc{
		o: &decoratedOsc{
			oscs:       []*osc{newOsc(true), newOsc(false)},
			adsr:       &adsr{tvalue: &transitiveValue{}},
			noteFilter: &noteFilter{filter: &filter{}},
			filter:     &filter{},
			formant:    newFormant(),
			lfos:       []*lfo{newLfo(), newLfo(), newLfo()},
			envelopes:  []*envelope{newEnvelope(), newEnvelope(), newEnvelope()},
		},
		activeNotes: make([]*noteOn, 0, 128),
		gain:        newTransitiveValue(),
	}
}

func (m *monoOsc) calc(
	events [][]*midiEvent,
	oscParams []*oscParams,
	adsrParams *adsrParams,
	noteFilterParams *noteFilterParams,
	filterParams *filterParams,
	formantParams *formantParams,
	lfoParams []*lfoParams,
	envelopeParams []*envelopeParams,
	velSense float64,
	glideTime int,
	echo *echo,
	out []float64,
) {
	m.o.applyParams(oscParams, adsrParams, noteFilterParams, filterParams, formantParams, lfoParams, envelopeParams)
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
					m.activeNotes[0] = data
					if len(m.activeNotes) == 1 {
						m.o.initWithNote(oscParams, data.note)
						gain := velocityToGain(data.velocity, velSense)
						m.gain.init(gain)
						event = enumNoteOn
					} else {
						m.o.glide(oscParams, m.activeNotes[0].note, glideTime)
						gain := velocityToGain(data.velocity, velSense)
						m.gain.exponential(float64(glideTime), gain, 0.001)
					}
				}
			case *noteOff:
				removed := 0
				for i := 0; i < len(m.activeNotes); i++ {
					if m.activeNotes[i].note == data.note {
						removed++
					} else {
						m.activeNotes[i-removed] = m.activeNotes[i]
					}
				}
				m.activeNotes = m.activeNotes[:len(m.activeNotes)-removed]
				if len(m.activeNotes) > 0 {
					m.o.glide(oscParams, m.activeNotes[0].note, glideTime)
					gain := velocityToGain(m.activeNotes[0].velocity, velSense)
					m.gain.exponential(float64(glideTime), gain, 0.001)
				} else {
					event = enumNoteOff
				}
			}
		}
		m.gain.step()
		value := m.o.step(event) * m.gain.value
		out[i] = echo.step(value)
	}
}
