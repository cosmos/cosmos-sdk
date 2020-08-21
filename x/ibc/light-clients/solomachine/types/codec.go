package types

import (
	"fmt"

	proto "github.com/gogo/protobuf/proto"
	tmcrypto "github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

// RegisterInterfaces register the ibc channel submodule interfaces to protobuf
// Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateClient{},
		&MsgUpdateClient{},
		&MsgSubmitClientMisbehaviour{},
	)
	registry.RegisterImplementations(
		(*clientexported.ClientState)(nil),
		&ClientState{},
	)
	registry.RegisterImplementations(
		(*clientexported.ConsensusState)(nil),
		&ConsensusState{},
	)
}

var (
	// SubModuleCdc references the global x/ibc/light-clients/solomachine module codec. Note, the codec
	// should ONLY be used in certain instances of tests and for JSON encoding..
	//
	// The actual codec used for serialization should be provided to x/ibc/light-clients/solomachine and
	// defined at the application level.
	SubModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
)

// PackPublicKey constructs a new Any packed with the given public key value. It returns
// an error if the public key can't be casted to a protobuf message or if the concrete
// implemention is not registered to the protobuf codec.
func PackPublicKey(publicKey tmcrypto.PubKey) (*codectypes.Any, error) {
	msg, ok := publicKey.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("cannot proto marshal %T", publicKey)
	}

	anyPublicKey, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}

	return anyPublicKey, nil
}
