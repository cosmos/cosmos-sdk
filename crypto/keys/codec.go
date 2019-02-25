package keys

import (
	usercryto "github.com/chain-dev/bschain/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	amino "github.com/tendermint/go-amino"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
)

var cdc = amino.NewCodec()

func init() {
	cryptoAmino.RegisterAmino(cdc)
	cdc.RegisterInterface((*Info)(nil), nil)
	cdc.RegisterConcrete(hd.BIP44Params{}, "crypto/keys/hd/BIP44Params", nil)
	cdc.RegisterConcrete(localInfo{}, "crypto/keys/localInfo", nil)
	cdc.RegisterConcrete(ledgerInfo{}, "crypto/keys/ledgerInfo", nil)
	cdc.RegisterConcrete(offlineInfo{}, "crypto/keys/offlineInfo", nil)
	cdc.RegisterConcrete(usercryto.PrivKeySm2{}, usercryto.Sm2PrivKeyAminoRoute, nil)
	cdc.RegisterConcrete(usercryto.PubKeySm2{}, usercryto.Sm2PubKeyAminoRoute, nil)
}
