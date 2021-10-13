package address

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Bech32Codec is a address.Codec based on bech32 encoding.
type Bech32Codec struct {
	Bech32Prefix string
}

var _ Codec = &Bech32Codec{}

// NewBech32Codec creates a new address.Codec based on bech32 encoding.
func NewBech32Codec(prefix string) Codec {
	return Bech32Codec{prefix}
}

// StringToBytes encodes text to bytes.
func (bc Bech32Codec) StringToBytes(text string) ([]byte, error) {
	hrp, bz, err := bech32.DecodeAndConvert(text)
	if err != nil {
		return nil, err
	}

	if hrp != bc.Bech32Prefix {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "hrp does not match bech32Prefix")
	}

	if err := sdk.VerifyAddressFormat(bz); err != nil {
		return nil, err
	}

	return bz, nil
}

// BytesToString decodes bytes to text.
func (bc Bech32Codec) BytesToString(bz []byte) (string, error) {
	text, err := bech32.ConvertAndEncode(bc.Bech32Prefix, bz)
	if err != nil {
		return "", err
	}

	return text, nil
}
