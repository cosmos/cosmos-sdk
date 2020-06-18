package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterCodec registers the Solo Machine types.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(ClientState{}, "ibc/client/solomachine/ClientState", nil)
	cdc.RegisterConcrete(ConsensusState{}, "ibc/client/solomachine/ConsensusState", nil)
	cdc.RegisterConcrete(Header{}, "ibc/client/solomachine/Header", nil)
	cdc.RegisterConcrete(Evidence{}, "ibc/client/solomachine/Evidence", nil)
	cdc.RegisterConcrete(MsgCreateClient{}, "ibc/client/solomachine/MsgCreateClient", nil)
	cdc.RegisterConcrete(MsgUpdateClient{}, "ibc/client/solomachine/MsgUpdateClient", nil)
	cdc.RegisterConcrete(MsgSubmitClientMisbehaviour{}, "ibc/client/solomachine/MsgSubmitClientMisbehaviour", nil)
	cdc.RegisterConcrete(cryptotypes.PublicKey{}, "cosmos/PublicKey", nil)
}

// RegisterInterfaces register the ibc channel submodule interfaces to protobuf
// Any.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateClient{},
		&MsgUpdateClient{},
		&MsgSubmitClientMisbehaviour{},
	)
}

var (
	amino = codec.New()

	// SubModuleCdc references the global x/ibc/06-solomachine module codec. Note, the codec
	// should ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/ibc/06-channel and
	// defined at the application level.
	SubModuleCdc = codec.NewHybridCodec(amino, cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(amino)
	amino.Seal()
}
