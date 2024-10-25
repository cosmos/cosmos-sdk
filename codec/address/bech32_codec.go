package address

import (
	"errors"
	"fmt"
	"strings"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type Bech32Codec struct {
	Bech32Prefix string
}

var _ address.Codec = &Bech32Codec{}

func NewBech32Codec(prefix string) address.Codec {
	return Bech32Codec{prefix}
}

// StringToBytes encodes text to bytes
func (bc Bech32Codec) StringToBytes(text string) ([]byte, error) {
	fmt.Println("StringToBytes for Bech32Codec", text)
	if len(strings.TrimSpace(text)) == 0 {
		return []byte{}, errors.New("empty address string is not allowed")
	}

	hrp, bz, err := bech32.DecodeAndConvert(text)
	fmt.Println("StringToBytes for Bech32Codec", hrp, bz, err)
	if err != nil {
		return nil, err
	}

	if hrp != bc.Bech32Prefix {
		return nil, errorsmod.Wrapf(sdkerrors.ErrLogic, "hrp does not match bech32 prefix: expected '%s' got '%s'", bc.Bech32Prefix, hrp)
	}

	err = sdk.VerifyAddressFormat(bz)
	fmt.Println("StringToBytes for Bech32Codec verification", err)
	if err != nil {
		return nil, err
	}

	return bz, nil
}

// BytesToString decodes bytes to text
func (bc Bech32Codec) BytesToString(bz []byte) (string, error) {
	if len(bz) == 0 {
		return "", nil
	}

	text, err := bech32.ConvertAndEncode(bc.Bech32Prefix, bz)
	if err != nil {
		return "", err
	}

	return text, nil
}
