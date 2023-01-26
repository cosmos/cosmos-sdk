package aminojson

import "github.com/cosmos/gogoproto/proto"

type JSONMarshaller interface {
	MarshalAmino(proto.Message) ([]byte, error)
}
