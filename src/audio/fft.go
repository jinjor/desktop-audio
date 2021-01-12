package audio

import (
	"log"
	"math"
	"math/cmplx"
)

// FFT ...
type FFT struct {
	bitReverseTable []int
	wTable          []complex128
	inverse         bool
}

// NewFFT ...
func NewFFT(length int, inverse bool) *FFT {
	return &FFT{
		bitReverseTable: makeBitReverseTable(length),
		wTable:          makeWTable(length),
	}
}
func makeBitReverseTable(n int) []int {
	array := make([]int, n)
	for i := 0; i < n; i++ {
		array[i] = bitReverse(i, n)
	}
	return array
}
func bitReverse(k, n int) int {
	m := 0
	for ; n > 1; n = n >> 1 {
		m = m<<1 + k&1
		k = k >> 1
	}
	return m
}
func makeWTable(n int) []complex128 {
	array := make([]complex128, n)
	w := -2.0 * math.Pi / float64(n)
	for i := 0; i < n; i++ {
		array[i] = cmplx.Exp(complex(0, w*float64(i)))
	}
	return array
}

// Calc ...
func (fft *FFT) Calc(x []complex128) {
	n := len(x)
	if n != len(fft.bitReverseTable) {
		log.Fatalf("length should be %v", len(fft.bitReverseTable))
	}
	for i := 0; i < n; i++ {
		rev := fft.bitReverseTable[i]
		if i < rev {
			x[i], x[rev] = x[rev], x[i]
		}
	}
	for m := 1; m < n; m = m << 1 {
		step := m << 1
		for k := 0; k < m; k++ {
			idx := n / step * k
			if fft.inverse {
				idx = n - idx
			}
			w := fft.wTable[idx]
			for i := k; i < n; i += step {
				j := i + m
				tmp := x[j] * w
				x[j] = x[i] - tmp
				x[i] = x[i] + tmp
			}
		}
	}
	if fft.inverse {
		for i := 0; i < n; i++ {
			x[i] /= complex(float64(n), 0)
		}
	}
}

// CalcReal ...
func (fft *FFT) CalcReal(x []float64) {
	n := len(x)
	cx := make([]complex128, n)
	for i := 0; i < n; i++ {
		cx[i] = complex(x[i], 0)
	}
	fft.Calc(cx)
	for i := 0; i < n; i++ {
		x[i] = real(cx[i])
	}
}

// CalcAbs ...
func (fft *FFT) CalcAbs(x []float64) {
	n := len(x)
	cx := make([]complex128, n)
	for i := 0; i < n; i++ {
		cx[i] = complex(x[i], 0)
	}
	fft.Calc(cx)
	for i := 0; i < n; i++ {
		x[i] = cmplx.Abs(cx[i])
	}
}

// HanningWindow ...
func HanningWindow(x []float64) {
	n := len(x)
	for i := 0; i < n; i++ {
		w := 0.5 - 0.5*math.Cos(2.0*math.Pi*float64(i)/float64(n))
		x[i] = x[i] * w
	}
}
