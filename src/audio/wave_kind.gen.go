// Code generated by gen/main.go; DO NOT EDIT.

package audio

const (
	waveNone = iota
	waveSine
	waveTriangle
	waveSquare
	waveSquareWT
	wavePulse
	waveSaw
	waveSawWT
	waveSawRev
	waveNoise
)

func waveKindFromString(s string) int {
	switch s {
	case "none":
		return waveNone
	case "sine":
		return waveSine
	case "triangle":
		return waveTriangle
	case "square":
		return waveSquare
	case "square-wt":
		return waveSquareWT
	case "pulse":
		return wavePulse
	case "saw":
		return waveSaw
	case "saw-wt":
		return waveSawWT
	case "saw-rev":
		return waveSawRev
	case "noise":
		return waveNoise
	}
	return waveNone
}
func waveKindToString(d int) string {
	switch d {
	case waveNone:
		return "none"
	case waveSine:
		return "sine"
	case waveTriangle:
		return "triangle"
	case waveSquare:
		return "square"
	case waveSquareWT:
		return "square-wt"
	case wavePulse:
		return "pulse"
	case waveSaw:
		return "saw"
	case waveSawWT:
		return "saw-wt"
	case waveSawRev:
		return "saw-rev"
	case waveNoise:
		return "noise"
	}
	return "none"
}
