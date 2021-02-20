package audio

import (
	"log"
	"math"
)

func makeNewSliceIfLengthsAreNotTheSame(slice []float64, expectedLength int) []float64 {
	if len(slice) == expectedLength {
		return slice
	}
	if cap(slice) >= expectedLength {
		return slice[:expectedLength]
	}
	return make([]float64, expectedLength)
}

func makeNoFilterH(feedforward []float64, feedback []float64) ([]float64, []float64) {
	feedforward = makeNewSliceIfLengthsAreNotTheSame(feedforward, 1)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 0)
	feedforward[0] = 1
	return feedforward, feedback
}

func makeFIRLowpassH(feedforward []float64, feedback []float64, N int, fc float64, windowFunc func(float64) float64) ([]float64, []float64) {
	w0 := 2 * math.Pi * fc
	if N%2 != 0 {
		log.Panicf("N should be even")
	}
	h := makeNewSliceIfLengthsAreNotTheSame(feedforward, N+1)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 0)
	for i := 0; i <= N; i++ {
		n := float64(i - N/2)
		h[i] = 2 * fc * sinc(w0*n)
	}
	applyWindow(h, windowFunc)
	return h, feedback
}

func makeFIRHighpassH(feedforward []float64, feedback []float64, N int, fc float64, windowFunc func(float64) float64) ([]float64, []float64) {
	w0 := 2 * math.Pi * fc
	if N%2 != 0 {
		log.Panicf("N should be even")
	}
	h := makeNewSliceIfLengthsAreNotTheSame(feedforward, N+1)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 0)
	for i := 0; i <= N; i++ {
		n := float64(i - N/2)
		h[i] = sinc(math.Pi*n) - 2*fc*sinc(w0*n)
	}
	applyWindow(h, windowFunc)
	return h, feedback
}

func makeBiquadLowpassH(feedforward []float64, feedback []float64, fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := (1 - math.Cos(w0)) / 2
	b1 := (1 - math.Cos(w0))
	b2 := (1 - math.Cos(w0)) / 2
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	feedforward = makeNewSliceIfLengthsAreNotTheSame(feedforward, 3)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 2)
	feedforward[0], feedforward[1], feedforward[2] = b0/a0, b1/a0, b2/a0
	feedback[0], feedback[1] = a1/a0, a2/a0
	return feedforward, feedback
}
func makeBiquadHighpassH(feedforward []float64, feedback []float64, fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := (1 + math.Cos(w0)) / 2
	b1 := -(1 + math.Cos(w0))
	b2 := (1 + math.Cos(w0)) / 2
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	feedforward = makeNewSliceIfLengthsAreNotTheSame(feedforward, 3)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 2)
	feedforward[0], feedforward[1], feedforward[2] = b0/a0, b1/a0, b2/a0
	feedback[0], feedback[1] = a1/a0, a2/a0
	return feedforward, feedback
}

func makeBiquadBandpass1H(feedforward []float64, feedback []float64, fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := math.Sin(w0) / 2
	b1 := 0.0
	b2 := -math.Sin(w0) / 2
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	feedforward = makeNewSliceIfLengthsAreNotTheSame(feedforward, 3)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 2)
	feedforward[0], feedforward[1], feedforward[2] = b0/a0, b1/a0, b2/a0
	feedback[0], feedback[1] = a1/a0, a2/a0
	return feedforward, feedback
}

func makeBiquadBandpass2H(feedforward []float64, feedback []float64, fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := alpha
	b1 := 0.0
	b2 := -alpha
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	feedforward = makeNewSliceIfLengthsAreNotTheSame(feedforward, 3)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 2)
	feedforward[0], feedforward[1], feedforward[2] = b0/a0, b1/a0, b2/a0
	feedback[0], feedback[1] = a1/a0, a2/a0
	return feedforward, feedback
}

func makeBiquadNotchH(feedforward []float64, feedback []float64, fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := 1.0
	b1 := -2 * math.Cos(w0)
	b2 := 1.0
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	feedforward = makeNewSliceIfLengthsAreNotTheSame(feedforward, 3)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 2)
	feedforward[0], feedforward[1], feedforward[2] = b0/a0, b1/a0, b2/a0
	feedback[0], feedback[1] = a1/a0, a2/a0
	return feedforward, feedback
}

func makeBiquadAllpassH(feedforward []float64, feedback []float64, fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := 1 - alpha
	b1 := -2 * math.Cos(w0)
	b2 := 1 + alpha
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	feedforward = makeNewSliceIfLengthsAreNotTheSame(feedforward, 3)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 2)
	feedforward[0], feedforward[1], feedforward[2] = b0/a0, b1/a0, b2/a0
	feedback[0], feedback[1] = a1/a0, a2/a0
	return feedforward, feedback
}

func makeBiquadPeakingEQH(feedforward []float64, feedback []float64, fc float64, q float64, dBgain float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	A := math.Pow(10, dBgain/40)
	b0 := 1 + alpha*A
	b1 := -2 * math.Cos(w0)
	b2 := 1 - alpha*A
	a0 := 1 + alpha/A
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha/A
	feedforward = makeNewSliceIfLengthsAreNotTheSame(feedforward, 3)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 2)
	feedforward[0], feedforward[1], feedforward[2] = b0/a0, b1/a0, b2/a0
	feedback[0], feedback[1] = a1/a0, a2/a0
	return feedforward, feedback
}

func makeBiquadLowShelfH(feedforward []float64, feedback []float64, fc float64, q float64, dBgain float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	A := math.Pow(10, dBgain/40)
	b0 := A * ((A + 1) - (A-1)*math.Cos(w0) + 2*math.Sqrt(A)*alpha)
	b1 := 2 * A * ((A - 1) - (A+1)*math.Cos(w0))
	b2 := A * ((A + 1) - (A-1)*math.Cos(w0) - 2*math.Sqrt(A)*alpha)
	a0 := (A + 1) + (A-1)*math.Cos(w0) + 2*math.Sqrt(A)*alpha
	a1 := -2 * ((A - 1) + (A+1)*math.Cos(w0))
	a2 := (A + 1) + (A-1)*math.Cos(w0) - 2*math.Sqrt(A)*alpha
	feedforward = makeNewSliceIfLengthsAreNotTheSame(feedforward, 3)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 2)
	feedforward[0], feedforward[1], feedforward[2] = b0/a0, b1/a0, b2/a0
	feedback[0], feedback[1] = a1/a0, a2/a0
	return feedforward, feedback
}

func makeBiquadHighShelfH(feedforward []float64, feedback []float64, fc float64, q float64, dBgain float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	A := math.Pow(10, dBgain/40)
	b0 := A * ((A + 1) + (A-1)*math.Cos(w0) + 2*math.Sqrt(A)*alpha)
	b1 := -2 * A * ((A - 1) + (A+1)*math.Cos(w0))
	b2 := A * ((A + 1) + (A-1)*math.Cos(w0) - 2*math.Sqrt(A)*alpha)
	a0 := (A + 1) - (A-1)*math.Cos(w0) + 2*math.Sqrt(A)*alpha
	a1 := 2 * ((A - 1) - (A+1)*math.Cos(w0))
	a2 := (A + 1) - (A-1)*math.Cos(w0) - 2*math.Sqrt(A)*alpha
	feedforward = makeNewSliceIfLengthsAreNotTheSame(feedforward, 3)
	feedback = makeNewSliceIfLengthsAreNotTheSame(feedback, 2)
	feedforward[0], feedforward[1], feedforward[2] = b0/a0, b1/a0, b2/a0
	feedback[0], feedback[1] = a1/a0, a2/a0
	return feedforward, feedback
}

func sinc(x float64) float64 {
	if math.Abs(x) < 0.000000001 {
		return 1
	}
	return math.Sin(x) / x
}
