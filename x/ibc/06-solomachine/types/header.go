package types

import (
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

var _ clientexported.Header = Header{}

// Header defines the Solo Machine consensus Header
type Header struct {
	Sequence     uint64
	Signature    Signature
	NewPublicKey PublicKey
}

// ClientType defines that the Header is a Solo Machine verficiation algorithm
func (Header) ClientType() clientexported.ClientType {
	return clientexported.SoloMachine
}

// GetHeight returns the current sequence number as the height
func (h Header) GetHeight() uint64 {
	return h.Sequence
}

// ValidateBasic
func (h Header) ValidateBasic() error {
	if h.Sequence == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "sequence number cannot be zero")
	}

	return nil
}
