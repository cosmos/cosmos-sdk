package bech32

import (
	errorsmod "cosmossdk.io/errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/btcutil/bech32"
)

// ConvertAndEncode converts from a base256 encoded byte string to base32 encoded byte string and then to bech32.
func ConvertAndEncode(hrp string, data []byte) (string, error) {
	converted, err := bech32.ConvertBits(data, 8, 5, true)
	if err != nil {
		return "", fmt.Errorf("encoding bech32 failed: %w", err)
	}

	return bech32.Encode(hrp, converted)
}

// DecodeAndConvert decodes a bech32 encoded string and converts to base256 encoded bytes.
func DecodeAndConvert(bech string) (string, []byte, error) {
	hrp, data, err := bech32.Decode(bech, 1023)
	if err != nil {
		return "", nil, fmt.Errorf("decoding bech32 failed: %w", err)
	}

	converted, err := bech32.ConvertBits(data, 5, 8, false)
	if err != nil {
		return "", nil, fmt.Errorf("decoding bech32 failed: %w", err)
	}

	return hrp, converted, nil
}

// VerifyAddressBech32 Verifies if a given address matches bech32 rules
func VerifyAddressBech32(bz []byte) error {
	if len(bz) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrUnknownAddress, "addresses cannot be empty")
	}

	if len(bz) > address.MaxAddrLen {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "address max length is %d, got %d", address.MaxAddrLen, len(bz))
	}

	return nil
}
