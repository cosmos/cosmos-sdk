package codec

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"
)

// TimeValue returns a ValueCodec for time.Time.
//
// Binary format: 8 bytes (big-endian) of int64 UnixNano.
// JSON format: RFC3339Nano string.
func TimeValue() ValueCodec[time.Time] {
	return timeValueCodec{}
}

type timeValueCodec struct{}

func (timeValueCodec) Encode(value time.Time) ([]byte, error) {
	value = value.UTC()
	n := value.UnixNano()

	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(n))
	return b, nil
}

func (timeValueCodec) Decode(b []byte) (time.Time, error) {
	if len(b) != 8 {
		return time.Time{}, fmt.Errorf("%w: invalid buffer size, wanted: 8", ErrEncoding)
	}

	n := int64(binary.BigEndian.Uint64(b))
	return time.Unix(0, n).UTC(), nil
}

func (timeValueCodec) EncodeJSON(value time.Time) ([]byte, error) {
	s := value.UTC().Format(time.RFC3339Nano)
	return json.Marshal(s)
}

func (timeValueCodec) DecodeJSON(b []byte) (time.Time, error) {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return time.Time{}, err
	}

	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}, err
	}

	return t.UTC(), nil
}

func (timeValueCodec) Stringify(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func (timeValueCodec) ValueType() string {
	return "time"
}

var _ ValueCodec[time.Time] = timeValueCodec{}
