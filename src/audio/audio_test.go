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

func benchmark(t *testing.T, audio *Audio, commands [][]string) {
	polyphony := 10
	times := 1000

	out := make([]byte, bufferSizeInBytes)
	for _, command := range commands {
		expectNoError(t, audio.update(command))
	}
	_, err := audio.Read(out)
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
	averageProcessTimeInMillis := (end - start) / float64(times) * 1000
	fmt.Printf("average process time: %.2fms\n", averageProcessTimeInMillis)
}

func TestBenchmark(t *testing.T) {
	audio, err := NewAudio("work/preset")
	expectNoError(t, err)
	defer expectNoError(t, audio.Close())

	fmt.Println("sine")
	benchmark(t, audio, [][]string{
		{"poly"},
		{"set", "osc", "0", "enabled", "true"},
		{"set", "osc", "1", "enabled", "true"},
		{"set", "filter", "enabled", "true"},
		{"set", "lfo", "0", "enabled", "true"},
		{"set", "lfo", "1", "enabled", "true"},
		{"set", "lfo", "2", "enabled", "true"},
		{"set", "envelope", "0", "enabled", "true"},
		{"set", "envelope", "1", "enabled", "true"},
		{"set", "envelope", "2", "enabled", "true"},
		{"set", "echo", "enabled", "true"},
	})

	fmt.Println("square-wt")
	benchmark(t, audio, [][]string{
		{"poly"},
		{"set", "osc", "0", "enabled", "true"},
		{"set", "osc", "1", "enabled", "true"},
		{"set", "filter", "enabled", "true"},
		{"set", "lfo", "0", "enabled", "true"},
		{"set", "lfo", "1", "enabled", "true"},
		{"set", "lfo", "2", "enabled", "true"},
		{"set", "envelope", "0", "enabled", "true"},
		{"set", "envelope", "1", "enabled", "true"},
		{"set", "envelope", "2", "enabled", "true"},
		{"set", "echo", "enabled", "true"},
		{"set", "osc", "0", "kind", "square-wt"},
		{"set", "osc", "1", "kind", "square-wt"},
	})
}
