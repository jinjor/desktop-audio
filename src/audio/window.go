package audio

import (
	"math"
)

func applyWindow(data []float64, windowFunc func(float64) float64) {
	n := len(data)
	for i := 0; i < n; i++ {
		x := float64(i) / float64(n-1)
		w := windowFunc(x)
		data[i] *= w
	}
}

func han(x float64) float64 {
	return 0.5 - 0.5*math.Cos(2.0*math.Pi*x)
}

func blackman(x float64) float64 {
	return 0.42 - 0.5*math.Cos(2.0*math.Pi*x) + 0.08*math.Cos(4.0*math.Pi*x)
}

func hamming(x float64) float64 {
	return 0.54 - 0.46*math.Cos(2.0*math.Pi*x)
}
