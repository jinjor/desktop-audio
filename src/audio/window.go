package audio

import (
	"math"
)

// ApplyWindow ...
func ApplyWindow(data []float64, windowFunc func(float64) float64) {
	n := len(data)
	for i := 0; i < n; i++ {
		x := float64(i) / float64(n-1)
		w := windowFunc(x)
		data[i] *= w
	}
}

// Han ...
func Han(x float64) float64 {
	return 0.5 - 0.5*math.Cos(2.0*math.Pi*x)
}

// Blackman ...
func Blackman(x float64) float64 {
	return 0.42 - 0.5*math.Cos(2.0*math.Pi*x) + 0.08*math.Cos(4.0*math.Pi*x)
}

// Hamming ...
func Hamming(x float64) float64 {
	return 0.54 - 0.46*math.Cos(2.0*math.Pi*x)
}
