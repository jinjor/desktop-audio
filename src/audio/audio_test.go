package audio

import (
	"fmt"
	"testing"
)

func expectNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("expected no error, but got: %v", err)
	}
}

func TestBenchmark(t *testing.T) {
	polyphony := 10
	times := 1000

	audio, err := NewAudio()
	expectNoError(t, err)
	defer expectNoError(t, audio.Close())
	out := make([]byte, bufferSizeInBytes)
	expectNoError(t, audio.update([]string{"poly"}))
	expectNoError(t, audio.update([]string{"set", "osc", "0", "enabled", "true"}))
	expectNoError(t, audio.update([]string{"set", "osc", "1", "enabled", "true"}))
	expectNoError(t, audio.update([]string{"set", "filter", "enabled", "true"}))
	expectNoError(t, audio.update([]string{"set", "lfo", "0", "enabled", "true"}))
	expectNoError(t, audio.update([]string{"set", "lfo", "1", "enabled", "true"}))
	expectNoError(t, audio.update([]string{"set", "lfo", "2", "enabled", "true"}))
	expectNoError(t, audio.update([]string{"set", "envelope", "0", "enabled", "true"}))
	expectNoError(t, audio.update([]string{"set", "envelope", "1", "enabled", "true"}))
	expectNoError(t, audio.update([]string{"set", "envelope", "2", "enabled", "true"}))
	expectNoError(t, audio.update([]string{"set", "echo", "enabled", "true"}))
	_, err = audio.Read(out)
	expectNoError(t, err)
	for n := 0; n < polyphony; n++ {
		audio.addMidiEvent(&noteOn{note: n})
	}
	start := now()
	for n := 0; n < times; n++ {
		_, err = audio.Read(out)
		expectNoError(t, err)
	}
	end := now()
	averageProcessTime := (end - start) / float64(times) * 1000
	fmt.Printf("average process time: %.2fms\n", averageProcessTime)
}
