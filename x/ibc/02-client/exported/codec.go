package exported

import (
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
)

// Codec defines the interface required to serialize custom x/ibc types.
type Codec interface {
	codec.Marshaler

	MarshalConnection(connectionexported.ConnectionI) ([]byte, error)
	UnmarshalConnection([]byte) (connectionexported.ConnectionI, error)
	MarshalChannel(channelexported.ChannelI) ([]byte, error)
	UnmarshalChannel([]byte) (channelexported.ChannelI, error)
}

// AnyCodec is an IBC Codec that uses google.protobuf.Any for IBC type encoding.
type AnyCodec struct {
	codec.Marshaler
}

// NewAnyCodec returns a new AnyCodec.
func NewAnyCodec(marshaler codec.Marshaler) Codec {
	return AnyCodec{Marshaler: marshaler}
}

// MarshalConnection marshals an Connection interface. If the given type implements
// the Marshaler interface, it is treated as a Proto-defined message and
// serialized that way. Otherwise, it falls back on the internal Amino codec.
func (c AnyCodec) MarshalConnection(connI connectionexported.ConnectionI) ([]byte, error) {
	return types.MarshalAny(connI)
}

// UnmarshalConnection returns a Connection interface from raw encoded evidence
// bytes of a Proto-based Connection type. An error is returned upon decoding
// failure.
func (c AnyCodec) UnmarshalConnection(bz []byte) (connectionexported.ConnectionI, error) {
	var connI connectionexported.ConnectionI
	if err := types.UnmarshalAny(c, &connI, bz); err != nil {
		return nil, err
	}

	return connI, nil
}

// MarshalConnectionJSON JSON encodes a connection object implementing the Connection
// interface.
func (c AnyCodec) MarshalConnectionJSON(connI connectionexported.ConnectionI) ([]byte, error) {
	msg, ok := connI.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("cannot proto marshal %T", connI)
	}

	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return c.MarshalJSON(any)
}

// UnmarshalConnectionJSON returns a Connection from JSON encoded bytes
func (c AnyCodec) UnmarshalConnectionJSON(bz []byte) (connectionexported.ConnectionI, error) {
	var any types.Any
	if err := c.UnmarshalJSON(bz, &any); err != nil {
		return nil, err
	}

	var connI connectionexported.ConnectionI
	if err := c.UnpackAny(&any, &connI); err != nil {
		return nil, err
	}

	return connI, nil
}

// MarshalChannel marshals an Channel interface. If the given type implements
// the Marshaler interface, it is treated as a Proto-defined message and
// serialized that way. Otherwise, it falls back on the internal Amino codec.
func (c AnyCodec) MarshalChannel(channI channelexported.ChannelI) ([]byte, error) {
	return types.MarshalAny(channI)
}

// UnmarshalChannel returns a Channel interface from raw encoded evidence
// bytes of a Proto-based Channel type. An error is returned upon decoding
// failure.
func (c AnyCodec) UnmarshalChannel(bz []byte) (channelexported.ChannelI, error) {
	var channI channelexported.ChannelI
	if err := types.UnmarshalAny(c, &channI, bz); err != nil {
		return nil, err
	}

	return channI, nil
}

// MarshalChannelJSON JSON encodes a channel object implementing the Channel
// interface.
func (c AnyCodec) MarshalChannelJSON(channI channelexported.ChannelI) ([]byte, error) {
	msg, ok := channI.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("cannot proto marshal %T", channI)
	}

	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return c.MarshalJSON(any)
}

// UnmarshalChannelJSON returns a Channel from JSON encoded bytes
func (c AnyCodec) UnmarshalChannelJSON(bz []byte) (channelexported.ChannelI, error) {
	var any types.Any
	if err := c.UnmarshalJSON(bz, &any); err != nil {
		return nil, err
	}

	var channI channelexported.ChannelI
	if err := c.UnpackAny(&any, &channI); err != nil {
		return nil, err
	}

	return channI, nil
}
