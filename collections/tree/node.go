package tree

import (
	"encoding/json"
	"strconv"
	"strings"

	"cosmossdk.io/collections/codec"
)

func NewNodeEncoder[K any](cdc codec.KeyCodec[K]) codec.ValueCodec[Node[K]] {
	return nodeCdc[K]{cdc}
}

type Node[K any] struct {
	Id     uint64
	Left   *uint64
	Right  *uint64
	Height uint64
	Key    K
}

type nodeCdc[K any] struct{ k codec.KeyCodec[K] }

func (n nodeCdc[K]) Encode(value Node[K]) ([]byte, error) {
	return json.Marshal(value)
}

func (n nodeCdc[K]) Decode(b []byte) (Node[K], error) {
	v := new(Node[K])
	return *v, json.Unmarshal(b, v)
}

func (n nodeCdc[K]) EncodeJSON(value Node[K]) ([]byte, error) {
	// TODO implement me
	panic("implement me")
}

func (n nodeCdc[K]) DecodeJSON(b []byte) (Node[K], error) {
	// TODO implement me
	panic("implement me")
}

func (n nodeCdc[K]) Stringify(value Node[K]) string {
	s := strings.Builder{}
	_ = s.WriteByte('{')
	_ = s.WriteByte(' ')
	_, _ = s.WriteString("ID: " + strconv.FormatUint(value.Id, 10))
	_, _ = s.WriteString(", ")
	_, _ = s.WriteString("Key: " + n.k.Stringify(value.Key))
	_ = s.WriteByte(',')
	_ = s.WriteByte(' ')
	_, _ = s.WriteString("Left: ")
	if value.Left == nil {
		_, _ = s.WriteString("<nil>")
	} else {
		_, _ = s.WriteString(strconv.FormatUint(*value.Left, 10))
	}
	_, _ = s.WriteString(", ")
	_, _ = s.WriteString("Right: ")
	if value.Right == nil {
		_, _ = s.WriteString("<nil>")
	} else {
		_, _ = s.WriteString(strconv.FormatUint(*value.Right, 10))
	}
	_, _ = s.WriteString(", ")
	_, _ = s.WriteString("Height: " + strconv.FormatUint(value.Height, 10))
	_, _ = s.WriteString("}")
	return s.String()
}

func (n nodeCdc[K]) ValueType() string {
	return "Node"
}
