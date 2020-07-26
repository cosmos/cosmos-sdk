package ledger

import (
	"github.com/KiraCore/cosmos-sdk/codec"
	cryptoAmino "github.com/KiraCore/cosmos-sdk/crypto/codec"
)

var cdc = codec.New()

func init() {
	RegisterAmino(cdc)
	cryptoAmino.RegisterCrypto(cdc)
}

// RegisterAmino registers all go-crypto related types in the given (amino) codec.
func RegisterAmino(cdc *codec.Codec) {
	cdc.RegisterConcrete(PrivKeyLedgerSecp256k1{},
		"tendermint/PrivKeyLedgerSecp256k1", nil)
}
