package grpc

import (
	"errors"
	"fmt"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/proto"

	_ "cosmossdk.io/api/amino" // Import amino.proto file for reflection
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
)

// protocdc defines the interface for marshaling and unmarshaling messages in server/v2
type protocdc interface {
	Marshal(v transaction.Msg) ([]byte, error)
	Unmarshal(data []byte, v transaction.Msg) error
	Name() string
}

type protoCodec struct {
	interfaceRegistry server.InterfaceRegistry
}

// newProtoCodec returns a reference to a new ProtoCodec
func newProtoCodec(interfaceRegistry server.InterfaceRegistry) *protoCodec {
	return &protoCodec{
		interfaceRegistry: interfaceRegistry,
	}
}

// Marshal implements BinaryMarshaler.Marshal method.
// NOTE: this function must be used with a concrete type which
// implements proto.Message. For interface please use the codec.MarshalInterface
func (pc *protoCodec) Marshal(o gogoproto.Message) ([]byte, error) {
	// Size() check can catch the typed nil value.
	if o == nil || gogoproto.Size(o) == 0 {
		// return empty bytes instead of nil, because nil has special meaning in places like store.Set
		return []byte{}, nil
	}

	return gogoproto.Marshal(o)
}

// Unmarshal implements BinaryMarshaler.Unmarshal method.
// NOTE: this function must be used with a concrete type which
// implements proto.Message. For interface please use the codec.UnmarshalInterface
func (pc *protoCodec) Unmarshal(bz []byte, ptr gogoproto.Message) error {
	err := gogoproto.Unmarshal(bz, ptr)
	if err != nil {
		return err
	}
	// err = codectypes.UnpackInterfaces(ptr, pc.interfaceRegistry) // TODO: identify if needed for grpc
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (pc *protoCodec) Name() string {
	return "cosmos-sdk-grpc-codec"
}

// GRPCCodec returns the gRPC Codec for this specific ProtoCodec
func (pc *protoCodec) GRPCCodec() encoding.Codec {
	return &grpcProtoCodec{cdc: pc}
}

// grpcProtoCodec is the implementation of the gRPC proto codec.
type grpcProtoCodec struct {
	cdc protocdc
}

var errUnknownProtoType = errors.New("codec: unknown proto type") // sentinel error

func (c grpcProtoCodec) Marshal(v any) ([]byte, error) {
	switch m := v.(type) {
	case proto.Message:
		protov2MarshalOpts := proto.MarshalOptions{Deterministic: true}
		return protov2MarshalOpts.Marshal(m)
	case gogoproto.Message:
		return c.cdc.Marshal(m)
	default:
		return nil, fmt.Errorf("%w: cannot marshal type %T", errUnknownProtoType, v)
	}
}

func (c grpcProtoCodec) Unmarshal(data []byte, v any) error {
	switch m := v.(type) {
	case proto.Message:
		return proto.Unmarshal(data, m)
	case gogoproto.Message:
		return c.cdc.Unmarshal(data, m)
	default:
		return fmt.Errorf("%w: cannot unmarshal type %T", errUnknownProtoType, v)
	}
}

func (c grpcProtoCodec) Name() string {
	return "cosmos-sdk-grpc-codec"
}
