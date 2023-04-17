package aminojson

import (
	"fmt"
	"io"
	"math"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	secondsName protoreflect.Name = "seconds"
	nanosName   protoreflect.Name = "nanos"
)

func marshalTimestamp(message protoreflect.Message, writer io.Writer) error {
	fields := message.Descriptor().Fields()
	secondsField := fields.ByName(secondsName)
	if secondsField == nil {
		return fmt.Errorf("expected seconds field")
	}

	nanosField := fields.ByName(nanosName)
	if nanosField == nil {
		return fmt.Errorf("expected nanos field")
	}

	seconds := message.Get(secondsField).Int()
	nanos := message.Get(nanosField).Int()
	if nanos < 0 {
		return fmt.Errorf("nanos must be non-negative on timestamp %v", message)
	}

	t := time.Unix(seconds, nanos).UTC()
	var str string
	if nanos == 0 {
		str = t.Format(time.RFC3339)
	} else {
		str = t.Format(time.RFC3339Nano)
	}

	_, err := fmt.Fprintf(writer, `"%s"`, str)
	return err
}

// MaxDurationSeconds the maximum number of seconds (when expressed as nanoseconds) which can fit in an int64.
// gogoproto encodes google.protobuf.Duration as a time.Duration, which is 64-bit signed integer.
const MaxDurationSeconds = int64(math.MaxInt64)/1e9 - 1

func marshalDuration(message protoreflect.Message, writer io.Writer) error {
	fields := message.Descriptor().Fields()
	secondsField := fields.ByName(secondsName)
	if secondsField == nil {
		return fmt.Errorf("expected seconds field")
	}

	// todo
	// check signs are consistent
	seconds := message.Get(secondsField).Int()
	if seconds > MaxDurationSeconds {
		return fmt.Errorf("%d seconds would overflow an int64 when represented as nanoseconds", seconds)
	}

	nanosField := fields.ByName(nanosName)
	if nanosField == nil {
		return fmt.Errorf("expected nanos field")
	}

	nanos := message.Get(nanosField).Int()
	totalNanos := nanos + (seconds * 1e9)
	_, err := writer.Write([]byte(fmt.Sprintf(`"%d"`, totalNanos)))
	return err
}
