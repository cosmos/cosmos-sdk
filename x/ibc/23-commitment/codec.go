package commitment

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers the necessary x/ibc/23-commitment interfaces and concrete types
// on the provided Amino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*RootI)(nil), nil)
	cdc.RegisterInterface((*PrefixI)(nil), nil)
	cdc.RegisterInterface((*PathI)(nil), nil)
	cdc.RegisterInterface((*ProofI)(nil), nil)

	cdc.RegisterConcrete(Root{}, "ibc/commitment/merkle/Root", nil)
	cdc.RegisterConcrete(Prefix{}, "ibc/commitment/merkle/Prefix", nil)
	cdc.RegisterConcrete(Path{}, "ibc/commitment/merkle/Path", nil)
	cdc.RegisterConcrete(Proof{}, "ibc/commitment/merkle/Proof", nil)
}

var (
	amino = codec.New()

	// SubModuleCdc references the global x/ibc/23-commitmentl module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/ibc/23-commitmentl and
	// defined at the application level.
	SubModuleCdc = codec.NewHybridCodec(amino)
)

func init() {
	RegisterCodec(amino)
	amino.Seal()
}
