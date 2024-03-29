package codec

import (
	"encoding/binary"
	"fmt"
	"time"
)

var timeSize = 8

type timeKey struct{}

func NewTimeKey() KeyCodec[time.Time] { return timeKey{} }

func (t timeKey) Encode(buffer []byte, key time.Time) (int, error) {
	if len(buffer) < timeSize {
		return 0, fmt.Errorf("buffer too small, required at least 8 bytes")
	}
	millis := key.UTC().UnixNano() / int64(time.Millisecond)
	binary.BigEndian.PutUint64(buffer, uint64(millis))
	return timeSize, nil
}

func (t timeKey) Decode(buffer []byte) (int, time.Time, error) {
	if len(buffer) != timeSize {
		return 0, time.Time{}, fmt.Errorf("invalid time buffer buffer size")
	}
	millis := int64(binary.BigEndian.Uint64(buffer))
	return timeSize, time.UnixMilli(millis).UTC(), nil
}

func (t timeKey) Size(_ time.Time) int { return timeSize }

func (t timeKey) EncodeJSON(value time.Time) ([]byte, error) { return value.MarshalJSON() }

func (t timeKey) DecodeJSON(b []byte) (time.Time, error) {
	time := time.Time{}
	err := time.UnmarshalJSON(b)
	return time, err
}

func (t timeKey) Stringify(key time.Time) string { return key.String() }
func (t timeKey) KeyType() string                { return "sdk/time.Time" }
func (t timeKey) EncodeNonTerminal(buffer []byte, key time.Time) (int, error) {
	return t.Encode(buffer, key)
}

func (t timeKey) DecodeNonTerminal(buffer []byte) (int, time.Time, error) {
	if len(buffer) < timeSize {
		return 0, time.Time{}, fmt.Errorf("invalid time buffer size, wanted: %d at least, got: %d", timeSize, len(buffer))
	}
	return t.Decode(buffer[:timeSize])
}
func (t timeKey) SizeNonTerminal(key time.Time) int { return t.Size(key) }
