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
func (fft *FFT) Calc(x []complex128) []complex128 {
	n := len(x)
	if n != len(fft.bitReverseTable) {
		log.Fatalf("length should be %v", len(fft.bitReverseTable))
	} else {
		tmp := make([]complex128, n)
		for i := 0; i < n; i++ {
			tmp[i] = x[fft.bitReverseTable[i]]
		}
		x = tmp
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
	return x
}

func w(n, k int) complex128 {
	return cmplx.Exp(complex(0, -2.0*math.Pi/float64(n)*float64(k)))
}

// CalcReal ...
func (fft *FFT) CalcReal(x []float64) []float64 {
	n := len(x)
	cx := make([]complex128, n)
	for i := 0; i < n; i++ {
		cx[i] = complex(x[i], 0)
	}
	cx = fft.Calc(cx)
	x = make([]float64, n)
	for i := 0; i < n; i++ {
		x[i] = real(cx[i])
	}
	return x
}

// CalcAbs ...
func (fft *FFT) CalcAbs(x []float64) []float64 {
	n := len(x)
	cx := make([]complex128, n)
	for i := 0; i < n; i++ {
		cx[i] = complex(x[i], 0)
	}
	cx = fft.Calc(cx)
	x = make([]float64, n)
	for i := 0; i < n; i++ {
		x[i] = cmplx.Abs(cx[i])
	}
	return x
}
