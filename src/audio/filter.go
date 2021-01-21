package audio

import (
	"log"
	"math"
)

func makeNoFilterH() ([]float64, []float64) {
	return []float64{1}, []float64{}
}

func makeFIRLowpassH(N int, fc float64, windowFunc func(float64) float64) ([]float64, []float64) {
	w0 := 2 * math.Pi * fc
	if N%2 != 0 {
		log.Panicf("N should be even")
	}
	h := make([]float64, N+1)
	for i := 0; i <= N; i++ {
		n := float64(i - N/2)
		h[i] = 2 * fc * sinc(w0*n)
	}
	applyWindow(h, windowFunc)
	return h, []float64{}
}

func makeFIRHighpassH(N int, fc float64, windowFunc func(float64) float64) ([]float64, []float64) {
	w0 := 2 * math.Pi * fc
	if N%2 != 0 {
		log.Panicf("N should be even")
	}
	h := make([]float64, N+1)
	for i := 0; i <= N; i++ {
		n := float64(i - N/2)
		h[i] = sinc(math.Pi*n) - 2*fc*sinc(w0*n)
	}
	applyWindow(h, windowFunc)
	return h, []float64{}
}

func makeBiquadLowpassH(fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := (1 - math.Cos(w0)) / 2
	b1 := (1 - math.Cos(w0))
	b2 := (1 - math.Cos(w0)) / 2
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	return []float64{b0 / a0, b1 / a0, b2 / a0}, []float64{a1 / a0, a2 / a0}
}

func makeBiquadHighpassH(fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := (1 + math.Cos(w0)) / 2
	b1 := -(1 + math.Cos(w0))
	b2 := (1 + math.Cos(w0)) / 2
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	return []float64{b0 / a0, b1 / a0, b2 / a0}, []float64{a1 / a0, a2 / a0}
}

func makeBiquadBandpass1H(fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := math.Sin(w0) / 2
	b1 := 0.0
	b2 := -math.Sin(w0) / 2
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	return []float64{b0 / a0, b1 / a0, b2 / a0}, []float64{a1 / a0, a2 / a0}
}

func makeBiquadBandpass2H(fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := alpha
	b1 := 0.0
	b2 := -alpha
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	return []float64{b0 / a0, b1 / a0, b2 / a0}, []float64{a1 / a0, a2 / a0}
}

func makeBiquadNotchH(fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := 1.0
	b1 := -2 * math.Cos(w0)
	b2 := 1.0
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	return []float64{b0 / a0, b1 / a0, b2 / a0}, []float64{a1 / a0, a2 / a0}
}

func makeBiquadAllpassH(fc float64, q float64) ([]float64, []float64) {
	// from RBJ's cookbook
	w0 := 2 * math.Pi * fc
	alpha := math.Sin(w0) / (2 * q)
	b0 := 1 - alpha
	b1 := -2 * math.Cos(w0)
	b2 := 1 + alpha
	a0 := 1 + alpha
	a1 := -2 * math.Cos(w0)
	a2 := 1 - alpha
	return []float64{b0 / a0, b1 / a0, b2 / a0}, []float64{a1 / a0, a2 / a0}
}

func makeBiquadPeakingEQH(fc float64, q float64, dBgain float64) ([]float64, []float64) {
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
	return []float64{b0 / a0, b1 / a0, b2 / a0}, []float64{a1 / a0, a2 / a0}
}

func makeBiquadLowShelfH(fc float64, q float64, dBgain float64) ([]float64, []float64) {
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
	return []float64{b0 / a0, b1 / a0, b2 / a0}, []float64{a1 / a0, a2 / a0}
}

func makeBiquadHighShelfH(fc float64, q float64, dBgain float64) ([]float64, []float64) {
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
	return []float64{b0 / a0, b1 / a0, b2 / a0}, []float64{a1 / a0, a2 / a0}
}

func sinc(x float64) float64 {
	if math.Abs(x) < 0.000000001 {
		return 1
	}
	return math.Sin(x) / x
}
