package module

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ProtoMsg interface {
	sdk.Msg
	codec.ProtoMarshaler
}

type InterfaceModule interface {
	Route() string
	// GetSigningProxy converts the provided message into another message type
	// that is to be used for signing or returns the original message
	// or an error if the msg type can't be handled.
	// Generally this is used with messages that have an interface member that
	// is encoded as a oneof at the app-level. The signing proxy should encode
	// this interface member using google.protobuf.Any to provide an app
	// independent representation for signatures
	GetSigningProxy(msg ProtoMsg) (ProtoMsg, error)

	// GetEncodingProxy converts the provided message into another message type
	// that is to be used for encoding or returns the original message
	// or an error if the msg type can't be handled.
	// Generally this is used with messages that have an interface member that
	// is encoded as a google.protobuf.Any in an app-independent (module level)
	// signing proxiy. The encoding proxy should encode
	// this interface member using an oneof at the app-level to minimize
	// transaction size over the wire and on disk
	GetEncodingProxy(msg ProtoMsg) (ProtoMsg, error)
}
