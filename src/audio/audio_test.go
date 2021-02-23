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
	defer expectNoError(t, audio.Close())
	expectNoError(t, err)
	out := make([]byte, bufferSizeInBytes)
	expectNoError(t, audio.update([]string{"poly"}))
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
