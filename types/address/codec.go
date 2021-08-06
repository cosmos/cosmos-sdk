package address

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type Codec interface {
	// ConvertAddressStringToBytes decodes text to bytes
	ConvertAddressStringToBytes(text string) ([]byte, error)
	// ConvertAddressBytesToString encodes bytes to text
	ConvertAddressBytesToString(bz []byte) (string, error)

}

// VerifyFormat checks validity of address format
func VerifyFormat(bz []byte) error {
	if len(bz) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrUnknownAddress, "addresses cannot be empty")
	}

	if len(bz) > MaxAddrLen {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "address max length is %d, got %d", MaxAddrLen, len(bz))
	}

	return nil
}
