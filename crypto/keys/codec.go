package keys

import (
	amino "github.com/tendermint/go-amino"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"

	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
)

var cdc = amino.NewCodec()

func init() {
	cryptoAmino.RegisterAmino(cdc)
	cdc.RegisterInterface((*Info)(nil), nil)
	cdc.RegisterConcrete(hd.BIP44Params{}, "crypto/keys/hd/BIP44Params", nil)
	cdc.RegisterConcrete(localInfo{}, "crypto/keys/localInfo", nil)
	cdc.RegisterConcrete(ledgerInfo{}, "crypto/keys/ledgerInfo", nil)
	cdc.RegisterConcrete(offlineInfo{}, "crypto/keys/offlineInfo", nil)
	cdc.RegisterConcrete(multiInfo{}, "crypto/keys/multiInfo", nil)
}
