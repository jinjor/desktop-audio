package audio

import (
	"log"
)

type polyOsc struct {
	// pooled + active = maxPoly
	pooled []*noteOsc
	active []*noteOsc
}

type noteOsc struct {
	*decoratedOsc
	note     int
	velocity int
}

func newPolyOsc() *polyOsc {
	pooled := make([]*noteOsc, maxPoly)
	for i := 0; i < len(pooled); i++ {
		pooled[i] = &noteOsc{
			decoratedOsc: newDecoratedOsc(),
		}
	}
	return &polyOsc{
		pooled: pooled,
	}
}
func (p *polyOsc) calc(
	events [][]*midiEvent,
	oscParams []*oscParams,
	adsrParams *adsrParams,
	noteFilterParams *noteFilterParams,
	filterParams *filterParams,
	formantParams *formantParams,
	lfoParams []*lfoParams,
	envelopeParams []*envelopeParams,
	velSense float64,
	echo *echo,
	out []float64,
) {
	for _, o := range p.active {
		o.applyParams(oscParams, adsrParams, noteFilterParams, filterParams, formantParams, lfoParams, envelopeParams)
	}
	for i := int64(0); i < int64(len(out)); i++ {
		events := events[i]
		for j := 0; j < len(events); j++ {
			switch data := events[j].event.(type) {
			case *noteOn:
				lenPooled := len(p.pooled)
				if lenPooled > 0 {
					o := p.pooled[lenPooled-1]
					p.pooled = p.pooled[:lenPooled-1]
					p.active = append(p.active, o)
					o.note = data.note
					o.velocity = data.velocity
					o.initWithNote(oscParams, data.note)
					o.applyParams(oscParams, adsrParams, noteFilterParams, filterParams, formantParams, lfoParams, envelopeParams)
				} else {
					log.Println("maxPoly exceeded")
				}
			}
		}
		out[i] = 0.0
		for _, o := range p.active {
			event := enumNoEvent
			for _, e := range events {
				switch data := e.event.(type) {
				case *noteOff:
					if data.note == o.note {
						event = enumNoteOff
					}
				case *noteOn:
					if data.note == o.note {
						event = enumNoteOn
					}
				}
			}
			gain := velocityToGain(o.velocity, velSense)
			out[i] += o.step(event) * gain
		}
		for j := len(p.active) - 1; j >= 0; j-- {
			o := p.active[j]
			if o.adsr.phase == phaseNone {
				p.active = append(p.active[:j], p.active[j+1:]...)
				p.pooled = append(p.pooled, o)
			}
		}
		out[i] = echo.step(out[i])
	}
}
