package stablejson

import (
	"io"
	"math"
	"strconv"
)

func marshalFloat(writer io.Writer, x float64) error {
	// PROTO3 SPEC:
	// JSON value will be a number or one of the special string values "NaN", "Infinity", and "-Infinity".
	// Either numbers or strings are accepted. Exponent notation is also accepted.
	// -0 is considered equivalent to 0.
	var err error
	if math.IsInf(x, -1) {
		_, err = writer.Write([]byte("-Infinity"))
	} else if math.IsInf(x, 1) {
		_, err = writer.Write([]byte("Infinity"))
	} else if math.IsNaN(x) {
		_, err = writer.Write([]byte("NaN"))
	} else {
		_, err = writer.Write([]byte(strconv.FormatFloat(x, 'f', -1, 64)))
	}
	return err
}
