package tx

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ValidateBasic performs stateless validation of the sign doc.
func (s *SignDocDirectAux) ValidateBasic() error {
	if len(s.BodyBytes) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("body bytes cannot be empty")
	}

	if s.PublicKey == nil {
		return sdkerrors.ErrInvalidPubKey.Wrap("public key cannot be empty")
	}

	if s.Tip != nil {
		if len(s.Tip.Amount) == 0 {
			return sdkerrors.ErrInvalidRequest.Wrap("tip amount cannot be empty")
		}

		if s.Tip.Tipper == "" {
			return sdkerrors.ErrInvalidRequest.Wrap("tipper cannot be empty")
		}
	}

	return nil
}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (s *SignDocDirectAux) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpacker.UnpackAny(s.PublicKey, new(cryptotypes.PubKey))
}

// ValidateBasic performs stateless validation of the auxiliary tx.
func (a *AuxTx) ValidateBasic() error {
	if len(a.Sig) == 0 {
		return sdkerrors.ErrNoSignatures.Wrap("signature cannot be empty")
	}

	return a.GetSignDoc().ValidateBasic()
}

// UnpackInterfaces implements the UnpackInterfaceMessages.UnpackInterfaces method
func (a *AuxTx) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return a.GetSignDoc().UnpackInterfaces(unpacker)
}
