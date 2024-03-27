package codec

import (
	"encoding/binary"
	"fmt"
	"time"
)

var timeSize = 8

type timeKey[T time.Time] struct{}

func NewTimeKey[T time.Time]() KeyCodec[T] { return timeKey[T]{} }

func (t timeKey[T]) Encode(buffer []byte, key T) (int, error) {
	if len(buffer) < timeSize {
		return 0, fmt.Errorf("buffer too small, required at least 8 bytes")
	}
	millis := time.Time(key).UnixNano() / int64(time.Millisecond)
	binary.BigEndian.PutUint64(buffer, uint64(millis))
	return timeSize, nil
}

func (t timeKey[T]) Decode(buffer []byte) (int, T, error) {
	if len(buffer) != timeSize {
		return 0, T{}, fmt.Errorf("invalid time buffer buffer size")
	}
	millis := int64(binary.BigEndian.Uint64(buffer))
	return timeSize, T(time.UnixMilli(millis).UTC()), nil
}

func (t timeKey[T]) Size(_ T) int { return timeSize }

func (t timeKey[T]) EncodeJSON(value T) ([]byte, error) { return time.Time(value).MarshalJSON() }

func (t timeKey[T]) DecodeJSON(b []byte) (T, error) {
	time := time.Time{}
	err := time.UnmarshalJSON(b)
	return T(time), err
}

func (t timeKey[T]) Stringify(key T) string { return time.Time(key).String() }
func (t timeKey[T]) KeyType() string        { return "sdk/time.Time" }
func (t timeKey[T]) EncodeNonTerminal(buffer []byte, key T) (int, error) {
	return t.Encode(buffer, key)
}

func (t timeKey[T]) DecodeNonTerminal(buffer []byte) (int, T, error) {
	if len(buffer) < timeSize {
		return 0, T{}, fmt.Errorf("invalid time buffer size, wanted: %d at least, got: %d", timeSize, len(buffer))
	}
	return t.Decode(buffer[:timeSize])
}
func (t timeKey[T]) SizeNonTerminal(key T) int { return t.Size(key) }
