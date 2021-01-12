package audio

import (
	"math"
	"testing"
)

func expectEqual(t *testing.T, actual, expected interface{}) {
	if actual != expected {
		t.Errorf("expected %v, but got: %v", expected, actual)
	}
}

func expectNearlyEqual(t *testing.T, actual, expected float64) {
	if math.Abs(actual-expected) > 0.0001 {
		t.Errorf("expected %v, but got: %v", expected, actual)
	}
}

func TestBitreverse(t *testing.T) {
	expectEqual(t, bitReverse(0, 8), 0)
	expectEqual(t, bitReverse(1, 8), 4)
	expectEqual(t, bitReverse(2, 8), 2)
	expectEqual(t, bitReverse(3, 8), 6)
	expectEqual(t, bitReverse(4, 8), 1)
	expectEqual(t, bitReverse(5, 8), 5)
	expectEqual(t, bitReverse(6, 8), 3)
	expectEqual(t, bitReverse(7, 8), 7)
}

func TestFFT(t *testing.T) {
	fft := NewFFT(8, false)
	x := []float64{0, 0.25, 0.5, 0.75, 1, 0.75, 0.5, 0.25}
	x = fft.CalcReal(x)
	expectNearlyEqual(t, x[0], 4)
	expectNearlyEqual(t, x[1], -(1 + math.Sqrt(2)/2))
	expectNearlyEqual(t, x[2], 0)
	expectNearlyEqual(t, x[3], -(1 - math.Sqrt(2)/2))
	expectNearlyEqual(t, x[4], 0)
	expectNearlyEqual(t, x[5], -(1 - math.Sqrt(2)/2))
	expectNearlyEqual(t, x[6], 0)
	expectNearlyEqual(t, x[7], -(1 + math.Sqrt(2)/2))
}
