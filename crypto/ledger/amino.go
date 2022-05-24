package ledger

import (
	"github.com/Stride-Labs/cosmos-sdk/codec"
	cryptoAmino "github.com/Stride-Labs/cosmos-sdk/crypto/codec"
)

var cdc = codec.NewLegacyAmino()

func init() {
	RegisterAmino(cdc)
	cryptoAmino.RegisterCrypto(cdc)
}

// RegisterAmino registers all go-crypto related types in the given (amino) codec.
func RegisterAmino(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(PrivKeyLedgerSecp256k1{},
		"tendermint/PrivKeyLedgerSecp256k1", nil)
}
