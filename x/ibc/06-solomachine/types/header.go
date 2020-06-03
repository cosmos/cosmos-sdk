package types

import (
	"github.com/tendermint/tendermint/crypto"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

var _ clientexported.Header = Header{}

// Header defines the Solo Machine consensus Header
type Header struct {
	Sequence  uint64        `json:"sequence" yaml:"sequence"`
	Signature []byte        `json:"signature" yaml:"signature"`
	NewPubKey crypto.PubKey `json:"new_pubkey" yaml:"new_pubkey"`
}

// ClientType defines that the Header is a Solo Machine verification algorithm.
func (Header) ClientType() clientexported.ClientType {
	return clientexported.SoloMachine
}

// GetHeight returns the current sequence number as the height.
func (h Header) GetHeight() uint64 {
	return h.Sequence
}

// ValidateBasic ensures that the sequence, signature and public key have all
// been initialized.
func (h Header) ValidateBasic() error {
	if h.Sequence == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "sequence number cannot be zero")
	}

	if len(h.Signature) == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "signature cannot be empty")
	}

	if h.NewPubKey == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "new public key is nil")
	}

	return nil
}
