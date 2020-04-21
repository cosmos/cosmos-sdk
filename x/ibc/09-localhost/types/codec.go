package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

const (
	// SubModuleName for the localhost (loopback) client
	SubModuleName = "localhost"
)

// RegisterCodec registers the necessary x/ibc/09-localhost interfaces and concrete types
// on the provided Amino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(ClientState{}, "ibc/client/localhost/ClientState", nil)
	cdc.RegisterConcrete(MsgCreateClient{}, "ibc/client/localhost/MsgCreateClient", nil)
}

var (
	amino = codec.New()

	// SubModuleCdc references the global x/ibc/07-tendermint module codec. Note, the codec
	// should ONLY be used in certain instances of tests and for JSON encoding as Amino
	// is still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/ibc/07-tendermint and
	// defined at the application level.
	SubModuleCdc = codec.NewHybridCodec(amino)
)

func init() {
	RegisterCodec(amino)
	codec.RegisterCrypto(amino)
	amino.Seal()
}
