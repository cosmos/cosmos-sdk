package signing

import (
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/client/context"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
)

type ClientSignature struct {
	pubKey    *cryptotypes.PublicKey
	signature []byte
	modeInfo  *types.ModeInfo
	codec     cryptotypes.PublicKeyCodec
}

type ModeInfoSignature interface {
	context.ClientSignature
	GetModeInfo() *types.ModeInfo
	SetModeInfo(modeInfo *types.ModeInfo)
}

func (c ClientSignature) GetPubKey() crypto.PubKey {
	pk, err := c.codec.Decode(c.pubKey)
	if err != nil {
		panic(err)
	}
	return pk
}

func (c ClientSignature) GetSignature() []byte {
	return c.signature
}

func (c *ClientSignature) SetPubKey(key crypto.PubKey) error {
	pk, err := c.codec.Encode(key)
	if err != nil {
		return err
	}
	c.pubKey = pk
	return nil
}

func (c *ClientSignature) SetSignature(bytes []byte) {
	c.signature = bytes
}

func (c *ClientSignature) GetModeInfo() *types.ModeInfo {
	return c.modeInfo
}

func (c *ClientSignature) SetModeInfo(modeInfo *types.ModeInfo) {
	c.modeInfo = modeInfo
}
