package audio

import (
	"math"
)

// Han ...
func Han(data []float64) {
	n := len(data)
	for i := 0; i < n; i++ {
		x := float64(i) / float64(n)
		w := 0.5 - 0.5*math.Cos(2.0*math.Pi*x)
		data[i] = data[i] * w
	}
}

// Blackman ...
func Blackman(data []float64) {
	n := len(data)
	for i := 0; i < n; i++ {
		x := float64(i) / float64(n)
		w := 0.42 - 0.5*math.Cos(2.0*math.Pi*x) + 0.08*math.Cos(4.0*math.Pi*x)
		data[i] = data[i] * w
	}
}

// Hamming ...
func Hamming(data []float64) {
	n := len(data)
	for i := 0; i < n; i++ {
		x := float64(i) / float64(n)
		w := 0.54 - 0.46*math.Cos(2.0*math.Pi*x)
		data[i] = data[i] * w
	}
}
