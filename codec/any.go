package codec

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

func MarshalAny(x interface{}) ([]byte, error) {
	evidenceProto, ok := x.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("can't proto marshal evidence")
	}
	any, err := types.MarshalAny(evidenceProto)
	if err != nil {
		return nil, err
	}
	return any.Marshal()
}

func UnmarshalAny(bz []byte) (interface{}, error) {
	any := types.Any{}
	err := any.Unmarshal(bz)
	if err != nil {
		return nil, err
	}
	dynAny := types.DynamicAny{}
	err = types.UnmarshalAny(&any, &dynAny)
	if err != nil {
		return nil, err
	}
	return dynAny.Message, nil
}
