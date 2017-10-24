package auth

import (
	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-crypto/keys"
	wire "github.com/tendermint/go-wire"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/util"
)

type oneSig struct {
	// Data is the payload
	Data util.RawTx
	// NamedSig holds credentials and exposes Sign
	*NamedSig
}

var _ Signable = oneSig{}
var _ sdk.Msg = oneSig{}
var _ keys.Signable = oneSig{}

func (o oneSig) SignBytes() []byte {
	return wire.BinaryBytes(o.Data)
}

func (o oneSig) TxBytes() ([]byte, error) {
	// if o.NamedSig.Empty() {
	// 	return nil, errors.ErrMissingSignature()
	// }
	return wire.BinaryBytes(o), nil
}

func (o oneSig) Signers() ([]crypto.PubKey, error) {
	return o.NamedSig.Signers(o.SignBytes())
}

func (o oneSig) GetTx() interface{} {
	return o.Data
}

func OneSig(data []byte) keys.Signable {
	return oneSig{
		Data:     util.NewRawTx(data),
		NamedSig: NewSig(),
	}
}

type multiSig struct {
	// Data is the payload
	Data util.RawTx
	// NamedSig holds credentials and exposes Sign
	*NamedSigs
}

var _ Signable = oneSig{}
var _ sdk.Msg = oneSig{}
var _ keys.Signable = oneSig{}

func (m multiSig) SignBytes() []byte {
	return wire.BinaryBytes(m.Data)
}

func (m multiSig) TxBytes() ([]byte, error) {
	// if m.NamedSigs.Empty() {
	// 	return nil, errors.ErrMissingSignature()
	// }
	return wire.BinaryBytes(m), nil
}

func (m multiSig) Signers() ([]crypto.PubKey, error) {
	return m.NamedSigs.Signers(m.SignBytes())
}

func (m multiSig) GetTx() interface{} {
	return m.Data
}

func MultiSig(data []byte) keys.Signable {
	return multiSig{
		Data:      util.NewRawTx(data),
		NamedSigs: NewMultiSig(),
	}
}
