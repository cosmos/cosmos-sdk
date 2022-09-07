package stablejson

import (
	"math"
	"strconv"
	"strings"
)

func marshalFloat(writer *strings.Builder, x float64) {
	// PROTO3 SPEC:
	// JSON value will be a number or one of the special string values "NaN", "Infinity", and "-Infinity".
	// Either numbers or strings are accepted. Exponent notation is also accepted.
	// -0 is considered equivalent to 0.
	if math.IsInf(x, -1) {
		writer.WriteString("-Infinity")
	} else if math.IsInf(x, 1) {
		writer.WriteString("Infinity")
	} else if math.IsNaN(x) {
		writer.WriteString("NaN")
	} else {
		writer.WriteString(strconv.FormatFloat(x, 'f', -1, 64))
	}
}
